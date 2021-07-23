package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/jszwec/csvutil"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const expectedScheme = "Bearer"

func getContextFromRequest(r *http.Request, authAPI auth.API) (context.Context, error) {
	var token string
	jwtBearer := r.Header.Get("authorization")
	if jwtBearer == "" {
		token = r.URL.Query().Get("token")
	} else {
		splits := strings.SplitN(jwtBearer, " ", 2)
		if len(splits) < 2 {
			return nil, fmt.Errorf("bad authorization string")

		}
		if !strings.EqualFold(splits[0], expectedScheme) {
			return nil, fmt.Errorf("request unauthenticated with %s", expectedScheme)
		}
		token = splits[1]
	}

	if token == "" {
		return nil, errors.New("missing jwt token in request")
	}

	// Communication context
	ctx := metadata.NewIncomingContext(
		r.Context(), metadata.Pairs(auth.Header(), fmt.Sprintf("%s %s", auth.Scheme(), token)),
	)

	// Authorize the context
	ctx, err := authAPI.AuthorizeFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize request: %v", err)
	}

	return ctx, nil
}

func getTimeRange(startTime, endTime string) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	// 2021-04-28T07:58:37.620Z
	// 2006-01-02T15:04:05.999999999Z07:00
	if endTime == "" {
		return &timestamppb.Timestamp{}, &timestamppb.Timestamp{}, nil
	}
	var (
		st  time.Time
		err error
	)
	if startTime == "" {
		st = time.Unix(0, 0)
	} else {
		st, err = time.Parse(time.RFC3339Nano, startTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse start time: %v", err)
		}
	}

	et, err := time.Parse(time.RFC3339Nano, endTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse end time: %v", err)
	}
	return timestamppb.New(st), timestamppb.New(et), nil
}

type accountsFilter struct {
	StartTime            string
	EndTime              string
	Filter               bool
	ShowActiveAccounts   bool
	ShowInactiveAccounts bool
	ShowBlockedAccounts  bool
	ShowMales            bool
	ShowFemales          bool
	FilterCreationDate   bool
	FilterAccountGroups  bool
	Groups               []string
	ProjectIds           []string
}

type accountInfo struct {
	AccountId    string
	Email        string
	Phone        string
	Names        string
	IdNumber     string
	AccountState string
	Group        string
	LastLogin    string
	CreatedAt    string
}

func downloadUsersHandler(accountAPI account.AccountAPIServer, authAPI auth.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Method must be POST
		if r.Method != http.MethodPost {
			http.Error(w, "only POST method allowed", http.StatusBadRequest)
			return
		}

		downloadFilter := &accountsFilter{}
		err := json.NewDecoder(r.Body).Decode(downloadFilter)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode filter in request body: %v", err), http.StatusBadRequest)
			return
		}

		ctx, err := getContextFromRequest(r, authAPI)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Must have role of administrator
		_, err = authAPI.AuthorizeGroup(ctx, authAPI.AdminGroups()...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		startTime, endTime, err := getTimeRange(downloadFilter.StartTime, downloadFilter.EndTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var (
			pageToken string
			pageSize  int32 = 1000
			next            = true
		)

		format := strings.ToLower(r.URL.Query().Get("format"))

		switch format {
		case "excel", "xlsx":
			// Excel file
			xlsxFile := excelize.NewFile()

			sheetName := xlsxFile.GetSheetName(xlsxFile.GetActiveSheetIndex())

			xlsxFile.SetSheetRow(sheetName, "A1", &[]interface{}{
				"AccountId", "Email", "Phone", "Names", "IdNumber", "Group", "AccountState", "LastLogin", "CreatedAt",
			})

			row := 1

			for next {
				listRes, err := accountAPI.ListAccounts(ctx, &account.ListAccountsRequest{
					PageToken: pageToken,
					PageSize:  pageSize,
					ListCriteria: &account.Criteria{
						Filter:               downloadFilter.Filter,
						ShowActiveAccounts:   downloadFilter.ShowActiveAccounts,
						ShowInactiveAccounts: downloadFilter.ShowInactiveAccounts,
						ShowBlockedAccounts:  downloadFilter.ShowBlockedAccounts,
						ShowMales:            downloadFilter.ShowMales,
						ShowFemales:          downloadFilter.ShowFemales,
						FilterCreationDate:   downloadFilter.FilterCreationDate,
						CreatedFrom:          startTime.Seconds,
						CreatedUntil:         endTime.Seconds,
						FilterAccountGroups:  downloadFilter.FilterAccountGroups,
						Groups:               downloadFilter.Groups,
						ProjectIds:           downloadFilter.ProjectIds,
					},
				})
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to list accounts: %v", err), http.StatusServiceUnavailable)
					return
				}

				pageToken = listRes.NextPageToken
				if listRes.NextPageToken == "" {
					next = false
				}

				row++

				for _, accountPB := range listRes.Accounts {
					xlsxFile.SetSheetRow(sheetName, fmt.Sprintf("A%d", row), &[]interface{}{
						accountPB.AccountId,
						accountPB.Email,
						accountPB.Phone,
						accountPB.Names,
						accountPB.IdNumber,
						accountPB.Group,
						accountPB.State.String(),
						accountPB.LastLogin,
						accountPB.CreatedAt,
					})
				}
			}

			// Set appropriate content type
			w.Header().Set("content-type", "text/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", "attachment; filename=accounts.xlsx")

			err = xlsxFile.Write(w)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create csv: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			// CSV Writer
			writer := csv.NewWriter(w)

			defer writer.Flush()

			// Decoder
			enc := csvutil.NewEncoder(writer)

			for next {
				listRes, err := accountAPI.ListAccounts(ctx, &account.ListAccountsRequest{
					PageToken: pageToken,
					PageSize:  pageSize,
					ListCriteria: &account.Criteria{
						Filter:               downloadFilter.Filter,
						ShowActiveAccounts:   downloadFilter.ShowActiveAccounts,
						ShowInactiveAccounts: downloadFilter.ShowInactiveAccounts,
						ShowBlockedAccounts:  downloadFilter.ShowBlockedAccounts,
						ShowMales:            downloadFilter.ShowMales,
						ShowFemales:          downloadFilter.ShowFemales,
						FilterCreationDate:   downloadFilter.FilterCreationDate,
						CreatedFrom:          startTime.Seconds,
						CreatedUntil:         endTime.Seconds,
						FilterAccountGroups:  downloadFilter.FilterAccountGroups,
						Groups:               downloadFilter.Groups,
						ProjectIds:           downloadFilter.ProjectIds,
					},
				})
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to list accounts: %v", err), http.StatusServiceUnavailable)
					return
				}

				pageToken = listRes.NextPageToken
				if listRes.NextPageToken == "" {
					next = false
				}

				for _, accountPB := range listRes.Accounts {
					err = enc.Encode(&accountInfo{
						AccountId: accountPB.AccountId,
						Email:     accountPB.Email,
						Phone:     accountPB.Phone,
						Names:     accountPB.Names,
						IdNumber:  accountPB.IdNumber,
						Group:     accountPB.Group,
						LastLogin: accountPB.LastLogin,
						CreatedAt: accountPB.CreatedAt,
					})
					if err != nil {
						http.Error(w, fmt.Sprintf("failed to encode ussd log to csv: %v", err), http.StatusInternalServerError)
						return
					}
				}
			}

			// Set appropriate content type
			w.Header().Set("content-type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=accounts.csv")

			if writer.Error() != nil && !errors.Is(writer.Error(), io.EOF) {
				http.Error(w, fmt.Sprintf("failed to create csv: %v", writer.Error()), http.StatusInternalServerError)
				return
			}
		}
	}
}

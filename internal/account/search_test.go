package account

import (
	"context"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro/pkg/grpc/auth"

	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Searching accounts @search", func() {
	var (
		searchReq *account.SearchAccountsRequest
		ctx       context.Context
	)

	BeforeEach(func() {
		searchReq = &account.SearchAccountsRequest{
			PageToken:      "",
			SearchCriteria: newCriteria(),
			View:           account.AccountView_FULL_VIEW,
			Query:          "test.",
		}
		ctx = context.Background()
	})

	Describe("Calling SearchAccounts with nil or malformed request", func() {
		It("should fail when request is nil", func() {
			searchReq = nil
			searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(searchRes).Should(BeNil())
		})
		It("should fail when page token is incorrect", func() {
			searchReq.PageToken = randomdata.RandStringRunes(48)
			searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(searchRes).Should(BeNil())
		})
	})

	Describe("Calling SearchAccounts with correct request payload", func() {
		Context("Calling search accounts with empty search query", func() {
			It("should succeed and no results returned", func() {
				searchReq.Query = ""
				searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(searchRes).ShouldNot(BeNil())
				Expect(len(searchRes.Accounts)).Should(BeZero())
			})
		})

		Context("Lets create one account first", func() {
			var names string
			It("should create the account without any error", func() {
				createReq := &account.CreateAccountRequest{
					Account:        fakeAccount(),
					PrivateAccount: fakePrivateAccount(),
					ProjectId:      "1",
				}
				createRes, err := AccountAPI.CreateAccount(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				names = createReq.Account.Names
			})

			Describe("Calling SearchAccounts", func() {
				It("should succeed", func() {
					searchReq.Query = names
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
				})
			})

			Describe("Calling SearchAccounts with date filter on", func() {
				It("should succeed", func() {
					searchReq.SearchCriteria.FilterCreationDate = true
					searchReq.SearchCriteria.CreatedFrom = time.Now().UnixNano()
					searchReq.SearchCriteria.CreatedUntil = time.Now().UnixNano() / 2
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
				})
			})

			Describe("Calling SearchAccounts with filter groups", func() {
				It("should succeed and returns only admin users", func() {
					searchReq.SearchCriteria.FilterAccountGroups = true
					searchReq.SearchCriteria.Groups = []string{auth.AdminGroup()}
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.Group).Should(BeElementOf(searchReq.SearchCriteria.Groups))
					}
				})
			})

			Describe("Calling SearchAccounts with show_males = true", func() {
				It("should succeed and returns only male users", func() {
					searchReq.SearchCriteria.ShowMales = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.Gender).Should(Equal(account.Account_MALE))
					}
				})
			})

			Describe("Calling SearchAccounts with show_females = true", func() {
				It("should succeed and returns only female users", func() {
					searchReq.SearchCriteria.ShowFemales = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.Gender).Should(Equal(account.Account_FEMALE))
					}
				})
			})

			Describe("Calling SearchAccounts with show_active_accounts true", func() {
				It("should succeed and returns only ACTIVE accounts", func() {
					searchReq.SearchCriteria.ShowActiveAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(Equal(account.AccountState_ACTIVE))
					}
				})
			})

			Describe("Calling SearchAccounts with show_inactive_accounts true", func() {
				It("should succeed and returns only INACTIVE accounts", func() {
					searchReq.SearchCriteria.ShowInactiveAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(Equal(account.AccountState_INACTIVE))
					}
				})
			})

			Describe("Calling SearchAccounts with show_blocked_accounts true", func() {
				It("should succeed and returns only BLOCKED accounts", func() {
					searchReq.SearchCriteria.ShowBlockedAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(Equal(account.AccountState_BLOCKED))
					}
				})
			})

			Describe("Calling SearchAccounts with show_blocked_accounts and show_active_accounts true", func() {
				It("should succeed and returns only BLOCKED accounts", func() {
					searchReq.SearchCriteria.ShowBlockedAccounts = true
					searchReq.SearchCriteria.ShowActiveAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					arr := []account.AccountState{
						account.AccountState_BLOCKED, account.AccountState_ACTIVE,
					}
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(BeElementOf(arr))
					}
				})
			})

			Describe("Calling SearchAccounts with show_inactive_accounts and show_active_accounts true", func() {
				It("should succeed and returns only ACTIVE or INACTIVE accounts", func() {
					searchReq.SearchCriteria.ShowInactiveAccounts = true
					searchReq.SearchCriteria.ShowActiveAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					arr := []account.AccountState{
						account.AccountState_INACTIVE, account.AccountState_ACTIVE,
					}
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(BeElementOf(arr))
					}
				})
			})

			Describe("Calling SearchAccounts with show_blocked_accounts and show_inactive_accounts true", func() {
				It("should succeed and returns only BLOCKED or INACTIVE accounts", func() {
					searchReq.SearchCriteria.ShowBlockedAccounts = true
					searchReq.SearchCriteria.ShowInactiveAccounts = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
					arr := []account.AccountState{
						account.AccountState_BLOCKED, account.AccountState_INACTIVE,
					}
					for _, adminPB := range searchRes.Accounts {
						Expect(adminPB.State).Should(BeElementOf(arr))
					}
				})
			})

			Describe("Calling SearchAccounts where all filters is true", func() {
				It("should succeed", func() {
					searchReq.SearchCriteria.ShowBlockedAccounts = true
					searchReq.SearchCriteria.ShowActiveAccounts = true
					searchReq.SearchCriteria.ShowInactiveAccounts = true
					searchReq.SearchCriteria.ShowFemales = true
					searchReq.SearchCriteria.ShowMales = true
					searchRes, err := AccountAPI.SearchAccounts(ctx, searchReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(searchRes).ShouldNot(BeNil())
				})
			})
		})
	})
})

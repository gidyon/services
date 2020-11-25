package settings

import (
	"context"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/settings"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Getting user settings", func() {
	var (
		getReq *settings.GetSettingsRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		getReq = &settings.GetSettingsRequest{
			OwnerId: fmt.Sprint(randomdata.Number(999, 9999)),
		}
		ctx = context.Background()
	})

	Describe("Getting settins with malformed request", func() {
		It("should fail when the request is nil", func() {
			getReq = nil
			getRes, err := SettingsAPI.GetSettings(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
		It("should fail when owner id is missing", func() {
			getReq.OwnerId = ""
			getRes, err := SettingsAPI.GetSettings(ctx, getReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(getRes).Should(BeNil())
		})
	})
})

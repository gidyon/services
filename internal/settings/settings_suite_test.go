package settings

import (
	"context"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro"
	"github.com/gidyon/micro/pkg/conn"
	"github.com/gidyon/services/pkg/api/settings"
	"github.com/gidyon/services/pkg/mocks"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestSettings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Settings Suite")
}

var (
	SettingsAPIServer *settingsAPIServer
	SettingsAPI       settings.SettingsAPIServer
)

const (
	dbAddress = "localhost:3306"
	schema    = "services"
)

func startDB() (*gorm.DB, error) {
	return conn.OpenGormConn(&conn.DBOptions{
		Dialect:  "mysql",
		Address:  "localhost:3306",
		User:     "root",
		Password: "hakty11",
		Schema:   schema,
	})
}

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start real database
	db, err := startDB()
	Expect(err).ShouldNot(HaveOccurred())

	logger := micro.NewLogger("settings_app")

	opt := &Options{
		SQLDB:         db,
		Logger:        logger,
		JWTSigningKey: []byte(randomdata.RandStringRunes(32)),
	}

	// Drop table
	Expect(db.Migrator().DropTable(settingsTable)).ShouldNot(HaveOccurred())

	// Auto migrate
	Expect(db.Migrator().AutoMigrate(&Model{})).ShouldNot(HaveOccurred())

	// Create settings API
	SettingsAPI, err = NewSettingsAPI(ctx, opt)
	Expect(err).ShouldNot(HaveOccurred())

	var ok bool
	SettingsAPIServer, ok = SettingsAPI.(*settingsAPIServer)
	Expect(ok).Should(BeTrue())

	SettingsAPIServer.authAPI = mocks.AuthAPI

	// Pasing incorrect payload
	_, err = NewSettingsAPI(nil, opt)
	Expect(err).Should(HaveOccurred())

	_, err = NewSettingsAPI(ctx, nil)
	Expect(err).Should(HaveOccurred())

	opt.SQLDB = nil
	_, err = NewSettingsAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.SQLDB = db
	opt.Logger = nil
	_, err = NewSettingsAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.Logger = logger
	opt.JWTSigningKey = nil
	_, err = NewSettingsAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.JWTSigningKey = []byte(randomdata.RandStringRunes(32))
	_, err = NewSettingsAPI(ctx, opt)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Expect(SettingsAPIServer.sqlDB.Close()).ShouldNot(HaveOccurred())
})

// Declarations for Ginkgo DSL
type Done ginkgo.Done
type Benchmarker ginkgo.Benchmarker

var GinkgoWriter = ginkgo.GinkgoWriter
var GinkgoRandomSeed = ginkgo.GinkgoRandomSeed
var GinkgoParallelNode = ginkgo.GinkgoParallelNode
var GinkgoT = ginkgo.GinkgoT
var CurrentGinkgoTestDescription = ginkgo.CurrentGinkgoTestDescription
var RunSpecs = ginkgo.RunSpecs
var RunSpecsWithDefaultAndCustomReporters = ginkgo.RunSpecsWithDefaultAndCustomReporters
var RunSpecsWithCustomReporters = ginkgo.RunSpecsWithCustomReporters
var Skip = ginkgo.Skip
var Fail = ginkgo.Fail
var GinkgoRecover = ginkgo.GinkgoRecover
var Describe = ginkgo.Describe
var FDescribe = ginkgo.FDescribe
var PDescribe = ginkgo.PDescribe
var XDescribe = ginkgo.XDescribe
var Context = ginkgo.Context
var FContext = ginkgo.FContext
var PContext = ginkgo.PContext
var XContext = ginkgo.XContext
var When = ginkgo.When
var FWhen = ginkgo.FWhen
var PWhen = ginkgo.PWhen
var XWhen = ginkgo.XWhen
var It = ginkgo.It
var FIt = ginkgo.FIt
var PIt = ginkgo.PIt
var XIt = ginkgo.XIt
var Specify = ginkgo.Specify
var FSpecify = ginkgo.FSpecify
var PSpecify = ginkgo.PSpecify
var XSpecify = ginkgo.XSpecify
var By = ginkgo.By
var Measure = ginkgo.Measure
var FMeasure = ginkgo.FMeasure
var PMeasure = ginkgo.PMeasure
var XMeasure = ginkgo.XMeasure
var BeforeSuite = ginkgo.BeforeSuite
var AfterSuite = ginkgo.AfterSuite
var SynchronizedBeforeSuite = ginkgo.SynchronizedBeforeSuite
var SynchronizedAfterSuite = ginkgo.SynchronizedAfterSuite
var BeforeEach = ginkgo.BeforeEach
var JustBeforeEach = ginkgo.JustBeforeEach
var JustAfterEach = ginkgo.JustAfterEach
var AfterEach = ginkgo.AfterEach

// Declarations for Gomega DSL
var RegisterFailHandler = gomega.RegisterFailHandler
var RegisterFailHandlerWithT = gomega.RegisterFailHandlerWithT
var RegisterTestingT = gomega.RegisterTestingT
var InterceptGomegaFailures = gomega.InterceptGomegaFailures
var Ω = gomega.Ω
var Expect = gomega.Expect
var ExpectWithOffset = gomega.ExpectWithOffset
var Eventually = gomega.Eventually
var EventuallyWithOffset = gomega.EventuallyWithOffset
var Consistently = gomega.Consistently
var ConsistentlyWithOffset = gomega.ConsistentlyWithOffset
var SetDefaultEventuallyTimeout = gomega.SetDefaultEventuallyTimeout
var SetDefaultEventuallyPollingInterval = gomega.SetDefaultEventuallyPollingInterval
var SetDefaultConsistentlyDuration = gomega.SetDefaultConsistentlyDuration
var SetDefaultConsistentlyPollingInterval = gomega.SetDefaultConsistentlyPollingInterval
var NewWithT = gomega.NewWithT
var NewGomegaWithT = gomega.NewGomegaWithT

// Declarations for Gomega Matchers
var Equal = gomega.Equal
var BeEquivalentTo = gomega.BeEquivalentTo
var BeIdenticalTo = gomega.BeIdenticalTo
var BeNil = gomega.BeNil
var BeTrue = gomega.BeTrue
var BeFalse = gomega.BeFalse
var HaveOccurred = gomega.HaveOccurred
var Succeed = gomega.Succeed
var MatchError = gomega.MatchError
var BeClosed = gomega.BeClosed
var Receive = gomega.Receive
var BeSent = gomega.BeSent
var MatchRegexp = gomega.MatchRegexp
var ContainSubstring = gomega.ContainSubstring
var HavePrefix = gomega.HavePrefix
var HaveSuffix = gomega.HaveSuffix
var MatchJSON = gomega.MatchJSON
var MatchXML = gomega.MatchXML
var MatchYAML = gomega.MatchYAML
var BeEmpty = gomega.BeEmpty
var HaveLen = gomega.HaveLen
var HaveCap = gomega.HaveCap
var BeZero = gomega.BeZero
var ContainElement = gomega.ContainElement
var BeElementOf = gomega.BeElementOf
var ConsistOf = gomega.ConsistOf
var HaveKey = gomega.HaveKey
var HaveKeyWithValue = gomega.HaveKeyWithValue
var BeNumerically = gomega.BeNumerically
var BeTemporally = gomega.BeTemporally
var BeAssignableToTypeOf = gomega.BeAssignableToTypeOf
var Panic = gomega.Panic
var BeAnExistingFile = gomega.BeAnExistingFile
var BeARegularFile = gomega.BeARegularFile
var BeADirectory = gomega.BeADirectory
var And = gomega.And
var SatisfyAll = gomega.SatisfyAll
var Or = gomega.Or
var SatisfyAny = gomega.SatisfyAny
var Not = gomega.Not
var WithTransform = gomega.WithTransform

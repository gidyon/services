package account

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"

	"github.com/gidyon/micro/v2"
	"github.com/gidyon/micro/v2/pkg/conn"
	micro_mock "github.com/gidyon/micro/v2/pkg/mocks"
	"github.com/gidyon/micro/v2/utils/encryption"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/mocks"
	redis "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Account Suite")
}

var (
	AccountAPI       account.AccountAPIServer
	AccountAPIServer *accountAPIServer
)

const (
	dbAddress    = "localhost:3306"
	schema       = "services"
	templatesDir = "/home/gideon/go/src/github.com/gidyon/services/internal/account/templates"
)

func startDB() (*gorm.DB, error) {
	return conn.OpenGormConn(&conn.DBOptions{
		Dialect:  "mysql",
		Address:  dbAddress,
		User:     "root",
		Password: "hakty11",
		Schema:   schema,
	})
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UnixNano())

	os.Setenv("MODE", "development")

	// Mysql database
	db, err := startDB()
	Expect(err).ShouldNot(HaveOccurred())

	Expect(db.Migrator().DropTable(accountsTable)).ShouldNot(HaveOccurred())
	Expect(db.Migrator().AutoMigrate(&Account{})).ShouldNot(HaveOccurred())

	// Redis db
	redisDB := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "localhost:6379",
	})

	// Logger
	Logger := micro.NewLogger("account service", 0)

	paginationHasher, err := encryption.NewHasher(string([]byte(randomdata.RandStringRunes(32))))
	Expect(err).ShouldNot(HaveOccurred())

	authAPI := micro_mock.AuthAPI

	opt := &Options{
		AppName:          "Accounts API",
		EmailDisplayName: "Accounts API",
		TemplatesDir:     templatesDir,
		ActivationURL:    randomdata.MacAddress(),
		PaginationHasher: paginationHasher,
		AuthAPI:          authAPI,
		SQLDBWrites:      db,
		SQLDBReads:       db,
		RedisDBWrites:    redisDB,
		RedisDBReads:     redisDB,
		Logger:           Logger,
		MessagingClient:  mocks.MessagingAPI,
		FirebaseAuth:     mocks.FirebaseAuthAPI,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	AccountAPI, err = NewAccountAPI(ctx, opt)
	Expect(err).ShouldNot(HaveOccurred())

	var ok bool
	AccountAPIServer, ok = AccountAPI.(*accountAPIServer)
	Expect(ok).Should(BeTrue())

	// When creating Accounts API with bad credentials
	_, err = NewAccountAPI(ctx, nil)
	Expect(err).Should(HaveOccurred())

	opt.SQLDBWrites = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.SQLDBWrites = db
	opt.MessagingClient = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.SQLDBReads = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.SQLDBReads = db
	opt.MessagingClient = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.MessagingClient = mocks.MessagingAPI
	opt.Logger = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.Logger = Logger
	opt.RedisDBWrites = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.RedisDBWrites = redisDB
	opt.PaginationHasher = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.PaginationHasher = paginationHasher
	opt.AuthAPI = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.AuthAPI = authAPI
	opt.RedisDBReads = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.RedisDBReads = redisDB
	opt.TemplatesDir = ""
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.TemplatesDir = templatesDir
	opt.AppName = ""
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.AppName = "Accounts API"
	opt.EmailDisplayName = ""
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.EmailDisplayName = "Accounts API"
	opt.ActivationURL = ""
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.ActivationURL = randomdata.MacAddress()
	opt.FirebaseAuth = nil
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).Should(HaveOccurred())

	opt.FirebaseAuth = mocks.FirebaseAuthAPI
	_, err = NewAccountAPI(ctx, opt)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Expect(AccountAPIServer.sqlDB.Close()).ShouldNot(HaveOccurred())
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

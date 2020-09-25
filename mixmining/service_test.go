package mixmining

import (
	"github.com/BorisBorshevsky/timemock"
	"github.com/nymtech/nym-directory/mixmining/mocks"
	"github.com/nymtech/nym-directory/models"
	. "github.com/onsi/ginkgo"
	"gotest.tools/assert"
)

// Some fixtures data to dry up tests a bit

func twoUpOneDown() []models.PersistedMixStatus {
	db := []models.PersistedMixStatus{}
	var status = newPersistedStatus()

	// IPv4
	// 2 ups and 1 down in last day
	status.PubKey = "key1"
	status.IPVersion = "4"
	status.Up = true

	status.Timestamp = minutesAgo(5)
	db = append(db, status)

	status.Timestamp = minutesAgo(10)
	db = append(db, status)

	status.Timestamp = minutesAgo(15)
	status.Up = false
	db = append(db, status)

	return db
}

func newPersistedStatus() models.PersistedMixStatus {
	mixStatus := newStatus()
	persisted := models.PersistedMixStatus{
		MixStatus: mixStatus,
		Timestamp: frozenNow(),
	}
	return persisted
}

func newStatus() models.MixStatus {
	return models.MixStatus{
		PubKey:    "key1",
		IPVersion: "4",
		Up:        false,
	}
}

func frozenNow() int64 {
	now := timemock.Now()
	timemock.Freeze(now) //time is frozen
	nanos := now.UnixNano()
	return nanos
}

var _ = Describe("mixmining.Service", func() {
	var mockDb mocks.IDb
	var status1 models.MixStatus
	var status2 models.MixStatus
	var persisted1 models.PersistedMixStatus
	var persisted2 models.PersistedMixStatus

	var serv Service
	var now = timemock.Now()
	timemock.Freeze(now)
	var frozen = timemock.Now().UnixNano()

	status1 = models.MixStatus{
		PubKey:    "key1",
		IPVersion: "4",
		Up:        true,
	}

	persisted1 = models.PersistedMixStatus{
		MixStatus: status1,
		Timestamp: frozen,
	}

	status2 = models.MixStatus{
		PubKey:    "key2",
		IPVersion: "6",
		Up:        true,
	}

	persisted2 = models.PersistedMixStatus{
		MixStatus: status2,
		Timestamp: frozen,
	}

	persistedList := []models.PersistedMixStatus{persisted1, persisted2}
	emptyList := []models.PersistedMixStatus{}

	Describe("Adding a mix status and creating a new summary report for a node", func() {
		Context("when no statuses have yet been saved", func() {
			It("should add a PersistedMixStatus to the db and save the new report", func() {
				mockDb = *new(mocks.IDb)
				serv = *NewService(&mockDb)
				mockDb.On("Add", persisted1)

				serv.CreateMixStatus(status1)
				mockDb.AssertCalled(GinkgoT(), "Add", persisted1)
			})
		})
	})
	Describe("Listing mix statuses", func() {
		Context("when receiving a list request", func() {
			It("should call to the Db", func() {
				mockDb = *new(mocks.IDb)
				serv = *NewService(&mockDb)
				mockDb.On("List", persisted1.PubKey, 1000).Return(persistedList)

				result := serv.List(persisted1.PubKey)

				mockDb.AssertCalled(GinkgoT(), "List", persisted1.PubKey, 1000)
				assert.Equal(GinkgoT(), persistedList[0].MixStatus.PubKey, result[0].MixStatus.PubKey)
				assert.Equal(GinkgoT(), persistedList[1].MixStatus.PubKey, result[1].MixStatus.PubKey)
			})
		})
	})

	Describe("Calculating uptime", func() {
		Context("when no statuses exist yet", func() {
			It("should return 0", func() {
				mockDb = *new(mocks.IDb)
				serv = *NewService(&mockDb)

				mockDb.On("ListDateRange", "key1", "4", frozenNow(), daysAgo(30)).Return(emptyList)

				uptime := serv.CalculateUptime(persisted1.PubKey, persisted1.IPVersion, daysAgo(30))
				assert.Equal(GinkgoT(), 0, uptime)
			})

		})
		Context("when 2 ups and 1 down exist in the given time period", func() {
			It("should return 66", func() {
				mockDb = *new(mocks.IDb)
				serv = *NewService(&mockDb)
				mockDb.On("ListDateRange", "key1", "4", frozenNow(), daysAgo(1)).Return(twoUpOneDown())

				uptime := serv.CalculateUptime("key1", "4", daysAgo(1))
				expected := 66 // percent
				assert.Equal(GinkgoT(), expected, uptime)
			})
		})
	})

	Describe("Saving a mix status report", func() {
		BeforeEach(func() {
			mockDb = *new(mocks.IDb)

		})
		Context("when no statuses exist yet", func() {

		})

		Context("when some statuses exist", func() {
			It("should save the report, leaving some fields blank", func() {
				expectedStatusReport := models.MixStatusReport{
					PubKey: persisted1.PubKey,
				}
				mockDb.On("SaveMixStatusReport", expectedStatusReport)
			})
		})

		Context("when all time periods exist", func() {

		})
	})
})

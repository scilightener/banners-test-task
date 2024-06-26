package tests

import (
	"net/url"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"avito-test-task/internal/models/dto/banner"
	"avito-test-task/tests/suit"
)

var (
	muTag         sync.Mutex
	muFeature     sync.Mutex
	lastTagID     int64 = 0
	lastFeatureID int64 = 0

	once               sync.Once
	expect             *httpexpect.Expect
	tokenUsr, tokenAdm string
)

func initTest(t *testing.T) (*httpexpect.Expect, string, string) {
	t.Helper()
	t.Parallel()
	once.Do(func() {
		t.Helper()

		s := suit.Setup(t)
		u := url.URL{
			Scheme: "http",
			Host:   s.Cfg.HTTPServer.Address,
		}

		expect = httpexpect.Default(t, u.String())

		user, err := s.JwtManager.GenerateToken("user")
		require.NoError(t, err)
		admin, err := s.JwtManager.GenerateToken("admin")
		require.NoError(t, err)
		tokenUsr, tokenAdm = user, admin
	})

	return expect, tokenUsr, tokenAdm
}

// getNextTagIDs returns the next tag IDs.
// It is unique for each call. It is thread-safe.
// The uniqueness is needed to avoid conflicts with the database trigger.
func getNextTagIDs(count int) []int64 {
	res := make([]int64, count)
	muTag.Lock()
	defer muTag.Unlock()
	for i := range count {
		lastTagID++
		res[i] = lastTagID
	}
	return res
}

// getNextFeatureID returns the next feature ID.
// It is unique for each call. It is thread-safe.
// The uniqueness is needed to avoid conflicts with the database trigger.
func getNextFeatureID() int64 {
	muFeature.Lock()
	defer muFeature.Unlock()
	lastFeatureID++
	return lastFeatureID
}

func getCreateBannerDTO() banner.CreateDTO {
	return banner.CreateDTO{
		TagIDs:    getNextTagIDs(2),
		FeatureID: getNextFeatureID(),
		Content: banner.CreateContent{
			Title: gofakeit.Word(),
			Text:  gofakeit.Word(),
			URL:   gofakeit.URL(),
		},
		IsActive: true,
	}
}

func getUpdateBannerDTO() *banner.UpdateDTO {
	tagIDs := getNextTagIDs(2)
	title := gofakeit.Word()
	text := gofakeit.Word()
	u := gofakeit.URL()
	featureID := getNextFeatureID()
	isActive := true

	return &banner.UpdateDTO{
		TagIDs:    &tagIDs,
		FeatureID: &featureID,
		Content: &banner.UpdateContent{
			Title: &title,
			Text:  &text,
			URL:   &u,
		},
		IsActive: &isActive,
	}
}

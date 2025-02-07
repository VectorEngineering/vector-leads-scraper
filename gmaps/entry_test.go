package gmaps_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/Vector/vector-leads-scraper/gmaps"
)

func createGoQueryFromFile(t *testing.T, path string) *goquery.Document {
	t.Helper()

	fd, err := os.Open(path)
	require.NoError(t, err)

	defer fd.Close()

	doc, err := goquery.NewDocumentFromReader(fd)
	require.NoError(t, err)

	return doc
}

func Test_EntryFromJSON(t *testing.T) {
	expected := gmaps.Entry{
		Link:       "https://www.google.com/maps/place/Kipriakon/data=!4m2!3m1!1s0x14e732fd76f0d90d:0xe5415928d6702b47!10m1!1e1",
		Title:      "Kipriakon",
		Category:   "Restaurant",
		Categories: []string{"Restaurant"},
		Address:    "Old port, Limassol 3042",
		OpenHours: map[string][]string{
			"Monday":    {"12:30–10 pm"},
			"Tuesday":   {"12:30–10 pm"},
			"Wednesday": {"12:30–10 pm"},
			"Thursday":  {"12:30–10 pm"},
			"Friday":    {"12:30–10 pm"},
			"Saturday":  {"12:30–10 pm"},
			"Sunday":    {"12:30–10 pm"},
		},
		WebSite:      "",
		Phone:        "25 101555",
		PlusCode:     "M2CR+6X Limassol",
		ReviewCount:  396,
		ReviewRating: 4.2,
		Latitude:     34.670595399999996,
		Longtitude:   33.042456699999995,
		Cid:          "16519582940102929223",
		Status:       "Closed ⋅ Opens 12:30\u202fpm Tue",
		ReviewsLink:  "https://search.google.com/local/reviews?placeid=ChIJDdnwdv0y5xQRRytw1ihZQeU&q=Kipriakon&authuser=0&hl=en&gl=CY",
		Thumbnail:    "https://lh5.googleusercontent.com/p/AF1QipP4Y7A8nYL3KKXznSl69pXSq9p2IXCYUjVvOh0F=w408-h408-k-no",
		Timezone:     "Asia/Nicosia",
		PriceRange:   "€€",
		DataID:       "0x14e732fd76f0d90d:0xe5415928d6702b47",
		Images: []gmaps.Image{
			{
				Title: "All",
				Image: "https://lh5.googleusercontent.com/p/AF1QipP4Y7A8nYL3KKXznSl69pXSq9p2IXCYUjVvOh0F=w298-h298-k-no",
			},
			{
				Title: "Latest",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNgMqyaQs2MqH1oiGC44eDcvudurxQfNb2RuDsd=w224-h298-k-no",
			},
			{
				Title: "Videos",
				Image: "https://lh5.googleusercontent.com/p/AF1QipPZbq8v8K8RZfvL6gZ_4Dw6qwNJ_MUxxOOfBo7h=w224-h398-k-no",
			},
			{
				Title: "Menu",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNhoFtPcaLCIhdN3GhlJ6sQIvdhaESnRG8nyeC8=w397-h298-k-no",
			},
			{
				Title: "Food & drink",
				Image: "https://lh5.googleusercontent.com/p/AF1QipMbu-iiWkE4DsXx3aI7nGaqyXJKbBYCrBXvzOnu=w298-h298-k-no",
			},
			{
				Title: "Vibe",
				Image: "https://lh5.googleusercontent.com/p/AF1QipOGg_vrD4bzkOre5Ly6CFXuO3YCOGfFxQ-EiEkW=w224-h398-k-no",
			},
			{
				Title: "Fried green tomatoes",
				Image: "https://lh5.googleusercontent.com/p/AF1QipOziHd2hqM1jnK9KfCGf1zVhcOrx8Bj7VdJXj0=w397-h298-k-no",
			},
			{
				Title: "French fries",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNJyq7nAlKtsxxbNy4PHUZOhJ0k7HPP8tTAlwcV=w397-h298-k-no",
			},
			{
				Title: "By owner",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNRE2R5k13zT-0WG4b6XOD_BES9-nMK04hlCMVV=w298-h298-k-no",
			},
			{
				Title: "Street View & 360°",
				Image: "https://lh5.googleusercontent.com/p/AF1QipMwkHP8GmDCSuwnWS7pYVQvtDWdsdk-CUwxtsXL=w224-h298-k-no-pi-23.425545-ya289.20517-ro-8.658787-fo100",
			},
		},
		OrderOnline: []gmaps.LinkSource{
			{
				Link:   "https://foody.com.cy/delivery/lemesos/to-kypriakon?utm_source=google&utm_medium=organic&utm_campaign=google_reserve_place_order_action",
				Source: "foody.com.cy",
			},
			{
				Link:   "https://wolt.com/en/cyp/limassol/restaurant/kypriakon?utm_source=googlemapreserved&utm_campaign=kypriakon",
				Source: "wolt.com",
			},
		},
		Owner: gmaps.Owner{
			ID:   "102769814432182832009",
			Name: "Kipriakon (Owner)",
			Link: "https://www.google.com/maps/contrib/102769814432182832009",
		},
		CompleteAddress: gmaps.Address{
			Borough:    "",
			Street:     "Old port",
			City:       "Limassol",
			PostalCode: "3042",
			State:      "",
			Country:    "CY",
		},
		ReviewsPerRating: map[int]int{
			1: 37,
			2: 16,
			3: 27,
			4: 60,
			5: 256,
		},
	}

	raw, err := os.ReadFile("../testdata/raw.json")
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)

	require.Len(t, entry.About, 10)

	for _, about := range entry.About {
		require.NotEmpty(t, about.ID)
		require.NotEmpty(t, about.Name)
		require.NotEmpty(t, about.Options)
	}

	entry.About = nil

	require.Len(t, entry.PopularTimes, 7)

	for k, v := range entry.PopularTimes {
		require.Contains(t,
			[]string{
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
				"Sunday",
			}, k)

		for _, traffic := range v {
			require.GreaterOrEqual(t, traffic, 0)
			require.LessOrEqual(t, traffic, 100)
		}
	}

	monday := entry.PopularTimes["Monday"]
	require.Equal(t, 100, monday[20])

	entry.PopularTimes = nil
	entry.UserReviews = nil

	require.Equal(t, expected, entry)
}

func Test_EntryFromJSON2(t *testing.T) {
	fnames := []string{
		"../testdata/panic.json",
		"../testdata/panic2.json",
	}
	for _, fname := range fnames {
		raw, err := os.ReadFile(fname)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		_, err = gmaps.EntryFromJSON(raw)
		require.NoError(t, err)
	}
}

func Test_EntryFromJSONRaw2(t *testing.T) {
	raw, err := os.ReadFile("../testdata/raw2.json")

	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entry, err := gmaps.EntryFromJSON(raw)

	require.NoError(t, err)
	require.Greater(t, len(entry.About), 0)
}

func Test_EntryFromJsonC(t *testing.T) {
	raw, err := os.ReadFile("../testdata/output.json")

	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entries, err := gmaps.ParseSearchResults(raw)

	require.NoError(t, err)

	for _, entry := range entries {
		fmt.Printf("%+v\n", entry)
	}
}

func Test_haversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "same point",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     0,
			expected: 0,
		},
		{
			name:     "known distance",
			lat1:     34.0522,
			lon1:     -118.2437,
			lat2:     40.7128,
			lon2:     -74.0060,
			expected: 3935746.25, // in meters
		},
		{
			name:     "antipodal points",
			lat1:     90,
			lon1:     0,
			lat2:     -90,
			lon2:     0,
			expected: 20015086.79, // in meters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := gmaps.HaversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			require.InDelta(t, tt.expected, dist, dist*0.0001) // 0.01% tolerance
		})
	}
}

func TestEntry_IsWebsiteValidForEmail(t *testing.T) {
	tests := []struct {
		name     string
		website  string
		expected bool
	}{
		{
			name:     "valid website",
			website:  "https://example.com",
			expected: true,
		},
		{
			name:     "empty website",
			website:  "",
			expected: false,
		},
		{
			name:     "social media website",
			website:  "https://facebook.com/business",
			expected: false,
		},
		{
			name:     "invalid url",
			website:  "not-a-url",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &gmaps.Entry{WebSite: tt.website}
			result := e.IsWebsiteValidForEmail()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestEntry_Validate(t *testing.T) {
	tests := []struct {
		name        string
		entry       *gmaps.Entry
		expectError bool
	}{
		{
			name: "valid entry",
			entry: &gmaps.Entry{
				Title:      "Test Place",
				Category:   "Restaurant",
				Address:    "123 Test St",
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			expectError: false,
		},
		{
			name: "missing title",
			entry: &gmaps.Entry{
				Category:   "Restaurant",
				Address:    "123 Test St",
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			expectError: true,
		},
		{
			name: "missing category",
			entry: &gmaps.Entry{
				Title:      "Test Place",
				Address:    "123 Test St",
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEntry_CsvHeaders(t *testing.T) {
	entry := &gmaps.Entry{}
	headers := entry.CsvHeaders()
	expectedHeaders := []string{
		"input_id", "link", "title", "category", "address",
		"open_hours", "popular_times", "website", "phone",
		"plus_code", "review_count", "review_rating", "reviews_per_rating",
		"latitude", "longitude", "cid", "status", "descriptions",
		"reviews_link", "thumbnail", "timezone", "price_range",
		"data_id", "images", "reservations", "order_online",
		"menu", "owner", "complete_address", "about", "user_reviews",
		"emails",
	}

	require.Equal(t, expectedHeaders, headers)
}

func TestEntry_CsvRow(t *testing.T) {
	entry := &gmaps.Entry{
		Title:        "Test Place",
		Category:     "Restaurant",
		Address:      "123 Test St",
		Latitude:     34.0522,
		Longtitude:   -118.2437,
		Phone:        "123-456-7890",
		WebSite:      "https://example.com",
		ReviewRating: 4.5,
		ReviewCount:  100,
		OpenHours:    map[string][]string{"Monday": {"9-5"}, "Tuesday": {"9-5"}},
		PopularTimes: map[string]map[int]int{
			"Monday": {
				0: 10,
				1: 20,
				2: 30,
			},
		},
		PriceRange: "$$",
		Status:     "Open",
		Emails:     []string{"test@example.com"},
	}

	row := entry.CsvRow()
	require.Len(t, row, len(entry.CsvHeaders()))
	require.Contains(t, row, entry.Title)
	require.Contains(t, row, entry.Category)
	require.Contains(t, row, entry.Address)
}

func TestEntry_isWithinRadius(t *testing.T) {
	tests := []struct {
		name     string
		entry    *gmaps.Entry
		lat      float64
		lon      float64
		radius   float64
		expected bool
	}{
		{
			name: "within radius",
			entry: &gmaps.Entry{
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			lat:      34.0522,
			lon:      -118.2437,
			radius:   1000.0, // 1km
			expected: true,
		},
		{
			name: "outside radius",
			entry: &gmaps.Entry{
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			lat:      40.7128,
			lon:      -74.0060,
			radius:   1000.0,
			expected: false,
		},
		{
			name: "exactly on radius",
			entry: &gmaps.Entry{
				Latitude:   34.0522,
				Longtitude: -118.2437,
			},
			lat:      34.0522,
			lon:      -118.2437,
			radius:   gmaps.HaversineDistance(34.0522, -118.2437, 34.0522, -118.2437),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsWithinRadius(tt.lat, tt.lon, tt.radius)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_filterAndSortEntriesWithinRadius(t *testing.T) {
	entries := []*gmaps.Entry{
		{
			Title:      "Near Place",
			Latitude:   34.0522,
			Longtitude: -118.2437,
		},
		{
			Title:      "Far Place",
			Latitude:   40.7128,
			Longtitude: -74.0060,
		},
		{
			Title:      "Very Near Place",
			Latitude:   34.0523,
			Longtitude: -118.2436,
		},
		{
			Title:      "Outside Radius",
			Latitude:   34.1000,
			Longtitude: -118.3000,
		},
	}

	centerLat := 34.0523
	centerLon := -118.2436

	filtered := gmaps.FilterAndSortEntriesWithinRadius(entries, centerLat, centerLon, 1000.0)
	
	// Verify length
	require.Len(t, filtered, 2, "Should only include places within radius")

	// Verify distances and order
	for i := 1; i < len(filtered); i++ {
		prevDist := gmaps.HaversineDistance(filtered[i-1].Latitude, filtered[i-1].Longtitude, centerLat, centerLon)
		currDist := gmaps.HaversineDistance(filtered[i].Latitude, filtered[i].Longtitude, centerLat, centerLon)
		require.LessOrEqual(t, prevDist, currDist, "Entries should be sorted by distance")
	}

	// Verify specific entries
	require.Equal(t, "Very Near Place", filtered[0].Title, "Closest entry should be first")
	require.Equal(t, "Near Place", filtered[1].Title, "Second closest entry should be second")
}

func Test_stringSliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "multiple values",
			input:    []string{"a", "b", "c"},
			expected: "a, b, c",
		},
		{
			name:     "single value",
			input:    []string{"a"},
			expected: "a",
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gmaps.StringSliceToString(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_stringify(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string input",
			input:    "test",
			expected: "test",
		},
		{
			name:     "integer input",
			input:    42,
			expected: "42",
		},
		{
			name:     "float input",
			input:    3.14,
			expected: "3.140000",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gmaps.Stringify(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

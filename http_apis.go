package query

import (
	"github.com/google/safebrowsing"
	"github.com/moldabekov/virusgotal/vt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
)

func querySafeBrowsing(q *HttpQuery) ([]QueryResult, error) {
	key := os.Getenv("SAFE_BROWSING_API_KEY")
	// setforspecialdomain.com
	usr, err := user.Current()

	sb, err := safebrowsing.NewSafeBrowser(safebrowsing.Config{
		APIKey:    key,
		DBPath:    path.Join(usr.HomeDir, ".awwwq", "safebrowsing"),
		ServerURL: safebrowsing.DefaultServerURL,
		// Logger:    os.Stderr,
		// ProxyURL:  *proxyFlag,
	})

	if err != nil {
		return []QueryResult{}, err
	}

	results := []QueryResult{}
	threats, err := sb.LookupURLs([]string{q.Target.Url})

	if err != nil {
		return results, err
	}

	for _, threat := range threats {
		for _, t := range threat {
			parts := []string{
				t.Pattern,
				t.ThreatDescriptor.ThreatType.String(),
				t.ThreatDescriptor.ThreatEntryType.String(),
			}
			results = append(results, &TextResult{"threat", strings.Join(parts, ",")})
		}
	}
	return results, err
}

func summarizeVt(report govt.UrlReport) map[string]string {
	detected := 0
	unrated := 0
	clean := 0
	detectedOn := []string{}

	for key, scan := range report.Scans {

		if scan.Detected {
			detected += 1
			detectedOn = append(detectedOn, key)
		}

		if scan.Result == "unrated site" {
			unrated += 1
		} else {
			clean += 1
		}
	}

	return map[string]string{
		"detected-by": strings.Join(detectedOn, ","), // @todo need better struct
		"detections":  strconv.Itoa(detected),
		"unrated":     strconv.Itoa(unrated),
		"clean":       strconv.Itoa(clean),
	}

}
func queryVirustotal(q *HttpQuery) ([]QueryResult, error) {
	vtotal, err := govt.New(govt.SetApikey(os.Getenv("VIRUS_TOTAL_API_KEY")))

	if err != nil {
		return []QueryResult{}, err
	}

	report, err := vtotal.GetUrlReport(q.Target.Url)

	if err != nil {
		return []QueryResult{}, err
	}

	results := []QueryResult{}

	for key, scan := range report.Scans {

		if scan.Detected {
			results = append(results, &TextResult{"virustotal", key + " - " + scan.Result})
		}
	}
	return results, nil
}

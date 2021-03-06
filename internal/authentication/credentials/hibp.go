package credentials

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	base                = "https://api.pwnedpasswords.com"
	pwnedHashesEndpoint = "/range"
	pwnedHashesURL      = base + pwnedHashesEndpoint + "/"
)

// IsPasswordPwned determines whether or not the password is weak.
func IsPasswordPwned(password string) (bool, error) {
	if len(password) == 0 {
		return true, nil
	}

	hash := sha1.Sum([]byte(password))
	hashStr := fmt.Sprintf("%X", hash)
	lookup := hashStr[0:5]
	match := hashStr[5:]

	log.Printf("Checking HIBP for hashes starting with '%s'.", lookup)

	resp, err := http.Get(pwnedHashesURL + lookup)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with HTTP %s.", resp.Status)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyStr := string(bodyText)
	lines := strings.Split(bodyStr, "\n")

	for _, line := range lines {
		separated := strings.Split(line, ":")
		if len(separated) != 2 {
			return false, fmt.Errorf("HIPB API returned malformed response: %s", line)
		}

		if separated[0] == match {
			return true, nil
		}
	}

	return false, nil
}

package security

import (
	"regexp"
	"sync"
	"time"
)

var (
	mu sync.RWMutex

	// bannedIPs are IP's banned for scanning web exploits
	bannedIPs = make(map[string]time.Time)

	// webExploits is a list of vulnerable url patterns
	webExploits = []string{
		".*\\.php",
		".*phpMyAdmin.*",
		".*\\/wp-admin\\/.*",
		".*\\/wp-content\\/.*",
		".*\\/mysqldumper.*",
		".*\\/cgi-bin\\/.*",
		".*\\/admin\\/mysql\\/.*",
	}

	// userAgents is a list of user agents from bad bots/crawlers/spiders/pen test tools
	userAgents = []string{
		".*acunetix.*",
		".*webshag.*",
		".*sqlmap.*",
		"Alligator",
		"AlphaBot",
		"Arachmo",
		"Arachnophilia",
		"ArchiveBot",
		"Arukereso",
		"AskQuickly",
		"Asterias",
		"Astute",
		"Attach",
		"Autonomy",
	}
)

// IsCrawler detects crawlers/spiders/bots by user agent, ip and url
func IsCrawler(url string, ip string, useragent string, banDuration time.Duration) bool {
	if banDuration == 0 {
		banDuration = time.Minute * 5
	}
	// check if ip is in ban time
	if status, banTime := getBannedIP(ip); status && banTime.Add(banDuration).After(time.Now()) {
		return true
	}

	// check if UA is in list of penetration tools
	if match := getMatch(useragent, userAgents); match {
		SetBannedIP(ip)
		return true
	}

	// check if requested url is in list of web exploits
	if match := getMatch(url, webExploits); match {
		SetBannedIP(ip)
		return true
	}
	return false
}

// geMatch is getter of UserAgent matched result
func getMatch(str string, list []string) bool {
	for _, check := range list {
		if match, _ := regexp.MatchString(check, str); match {
			return true
		}
	}
	return false
}

// getBannedIP checks if IP registered
func getBannedIP(ip string) (bool, time.Time) {
	mu.RLock()
	defer mu.RUnlock()

	if banTime, ok := bannedIPs[ip]; ok {
		return true, banTime
	}
	return false, time.Time{}
}

// SetBannedIP registers IP to banned list
func SetBannedIP(ip string) {
	mu.Lock()
	defer mu.Unlock()

	bannedIPs[ip] = time.Now()
}

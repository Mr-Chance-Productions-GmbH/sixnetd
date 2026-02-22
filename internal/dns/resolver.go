package dns

import (
	"fmt"
	"os"
)

const resolverDir = "/etc/resolver"

// Write creates /etc/resolver/<domain> pointing at server.
// macOS mDNSResponder picks this up automatically — no restart needed.
func Write(domain, server string) error {
	content := fmt.Sprintf("nameserver %s\n", server)
	return os.WriteFile(resolverDir+"/"+domain, []byte(content), 0644)
}

// Remove deletes /etc/resolver/<domain>. A missing file is not an error.
func Remove(domain string) error {
	err := os.Remove(resolverDir + "/" + domain)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

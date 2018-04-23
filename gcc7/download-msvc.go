package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/davecgh/go-spew/spew"
)

type itemPayload struct {
	FileName string `json:"fileName"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	URL      string
}

type channelItem struct {
	ID       string         `json:"id"`
	Version  string         `json:"version"`
	Type     string         `json:"type"`
	Payloads []*itemPayload `json:"payloads"`
}

type manifestPackage struct {
	ID            string                      `json:"id"`
	Version       string                      `json:"version"`
	Type          string                      `json:"type"`
	Chip          string                      `json:"chip"`
	Language      string                      `json:"language"`
	Dependencies  map[string]dependencyFilter `json:"dependencies"`
	Payloads      []*itemPayload              `json:"payloads"`
	InstallParams struct {
		FileName   string `json:"fileName"`
		Parameters string `json:"parameters"`
	} `json:"installParams"`
	MSIProperties map[string]string `json:"msiProperties"`
}

type dependencyFilter struct {
	Version   string   `json:"version"`
	Type      string   `json:"type"`
	Chip      string   `json:"chip"`
	When      []string `json:"when"`
	Behaviors string   `json:"behaviors"`
}

func (f *dependencyFilter) UnmarshalJSON(b []byte) error {
	var format1 string
	type noMethods dependencyFilter
	var format2 noMethods

	if err := json.Unmarshal(b, &format1); err == nil {
		*f = dependencyFilter{}
		f.Version = format1
		return nil
	}
	if err := json.Unmarshal(b, &format2); err != nil {
		return err
	}

	*f = dependencyFilter(format2)
	return nil
}

var installedPackages = map[string]bool{
	"Microsoft.Net.4.6.1.FullRedist.NonThreshold": true,
	"Microsoft.Net.4.6.1.FullRedist.Threshold":    true,
	"Microsoft.VisualCpp.Redist.14":               true,
	"Microsoft.VisualCpp.Redist.14.Latest":        true,
}

func main() {
	log.SetOutput(os.Stdout)

	var channel struct {
		Items []*channelItem `json:"channelItems"`
	}
	doRequest("https://aka.ms/vs/15/release/channel", &channel)

	manifestItem := findChannelItem(channel.Items, "Microsoft.VisualStudio.Manifests.VisualStudio")

	var manifest struct {
		Packages []*manifestPackage `json:"packages"`
	}
	doRequest(manifestItem.Payloads[0].URL, &manifest)

	installPackage(manifest.Packages, "Microsoft.VisualStudio.Product.BuildTools", nil)
	installPackage(manifest.Packages, "Microsoft.VisualStudio.Workload.VCTools", nil)
	installPackage(manifest.Packages, "Microsoft.VisualStudio.Component.VC.140", nil)
	installPackage(manifest.Packages, "Microsoft.VisualStudio.Component.WinXP", nil)
	log.Println("Done!")
}

func doRequest(url string, v interface{}) {
	log.Println("Requesting URL:", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	validateSignature(data)

	log.Println("Decoding JSON data...")
	err = json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
}

func findChannelItem(items []*channelItem, id string) *channelItem {
	log.Println("Searching for channel item:", id)
	for _, i := range items {
		if i.ID == id {
			log.Println("Found", i.Type, "version", i.Version)
			return i
		}
	}
	panic("Could not find channel item: " + id)
}

func findManifestPackage(packages []*manifestPackage, id string, filter *dependencyFilter) *manifestPackage {
	anyWereWrongLanguage := false
	allWereWrongLanguage := true
	log.Println("Searching the manifest for package:", id)
	for _, p := range packages {
		if strings.EqualFold(p.ID, id) {
			if p.Language != "" && !strings.EqualFold(p.Language, "en-us") && p.Language != "neutral" {
				anyWereWrongLanguage = true
				log.Println("Skipping", p.Type, "version", p.Version, "as it has language:", p.Language)
				continue
			}
			allWereWrongLanguage = false
			if (filter == nil || filter.Chip == "") && p.Chip != "" && !strings.EqualFold(p.Chip, "x64") && p.Chip != "neutral" {
				log.Println("Skipping", p.Type, "version", p.Version, "as it has chip type", p.Chip)
				continue
			}
			if filter != nil && filter.Chip != "" && !strings.EqualFold(p.Chip, filter.Chip) {
				log.Println("Skipping", p.Type, "version", p.Version, "as it has chip type", p.Chip, "and", filter.Chip, "was expected.")
				continue
			}
			log.Println("Found", p.Type, "version", p.Version)
			return p
		}
	}
	if filter != nil && filter.Behaviors == "IgnoreApplicabilityFailures" {
		return nil
	}
	if anyWereWrongLanguage && allWereWrongLanguage {
		return nil
	}
	panic("Could not find package in the manifest: " + id)
}

func downloadPayload(dir string, payload *itemPayload) {
	log.Println("Downloading file", payload.FileName, "from URL:", payload.URL)

	resp, err := http.Get(payload.URL)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	f, err := os.Create(filepath.Join(dir, payload.FileName))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	hash := sha256.New()

	_, err = io.Copy(io.MultiWriter(hash, f), resp.Body)
	if err != nil {
		panic(err)
	}

	expected, err := hex.DecodeString(payload.SHA256)
	if err != nil {
		panic(err)
	}

	if actual := hash.Sum(nil); !bytes.Equal(expected, actual) {
		panic("Hash mismatch:\n  expected " + payload.SHA256 + "\n  actual " + hex.EncodeToString(actual))
	}
}

func installPackage(packages []*manifestPackage, id string, filter *dependencyFilter) {
	if installedPackages[id] {
		log.Println("Skipping already-installed package:", id)
		return
	}
	installedPackages[id] = true
	pkg := findManifestPackage(packages, id, filter)
	if pkg == nil {
		log.Println("Ignoring lack of applicable package:", id)
		return
	}

	for d, filter := range pkg.Dependencies {
		if filter.Type != "" {
			log.Println("Skipping", filter.Type, "package:", d)
			continue
		}
		if len(filter.When) != 0 {
			found := false
			for _, w := range filter.When {
				if w == "Microsoft.VisualStudio.Product.BuildTools" {
					found = true
					break
				}
			}
			if !found {
				log.Println("Skipping non-BuildTools package:", d)
				continue
			}
		}
		installPackage(packages, d, &filter)
	}

	if len(pkg.Payloads) == 0 {
		log.Println("Package has no payloads:", id)
		return
	}

	log.Println("Downloading payloads for package", id, "version", pkg.Version)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Println("Could not remove temp dir for", id, err)
		}
	}()

	for _, p := range pkg.Payloads {
		downloadPayload(dir, p)
	}

	switch pkg.Type {
	case "Exe":
		if pkg.InstallParams.FileName != "[Payload]" {
			panic("unexpected EXE install filename")
		}
		execCmd(dir, filepath.Join(dir, pkg.Payloads[0].FileName), splitParameters(replacePlaceholders(pkg.InstallParams.Parameters))...)
	case "Msi":
		args := []string{"/i", pkg.Payloads[0].FileName}
		for k, v := range pkg.MSIProperties {
			args = append(args, k+"="+replacePlaceholders(v))
		}
		execCmd(dir, "msiexec.exe", args...)
	case "Msu":
		execCmd(dir, "wusa.exe", pkg.Payloads[0].FileName, "/quiet", "/norestart")
	case "Vsix":
		extractVSIX(dir, pkg.Payloads[0].FileName)
	default:
		panic("Don't know how to install package type: " + pkg.Type)
	}
}

func execCmd(dir, program string, args ...string) {
	log.Println("Executing program", program, "with arguments", args)

	cmd := exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			switch ee.Sys().(syscall.WaitStatus).ExitStatus() {
			case 3010:
				log.Println("Ignoring exit code 3010: restart requested")
				return
			}
		}
		panic(err)
	}
}

func extractVSIX(dir, name string) {
	log.Println("Extracting VSIX file:", name)

	r, err := zip.OpenReader(filepath.Join(dir, name))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	for _, f := range r.File {
		path := filepath.ToSlash(filepath.Clean(f.Name))
		if !strings.HasPrefix(path, "Contents/") {
			continue
		}

		extractVSIXFile(filepath.Join("C:\\BuildTools", strings.TrimPrefix(path, "Contents/")), f)
	}
}

func extractVSIXFile(destPath string, f *zip.File) {
	if f.FileInfo().IsDir() {
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			panic(err)
		}

		return
	}

	r, err := f.Open()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		panic(err)
	}

	w, err := os.Create(destPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = io.Copy(w, r)
	if err != nil {
		panic(err)
	}
}

func replacePlaceholders(source string) string {
	source = strings.Replace(source, "[CEIPConsent]", "/CEIPConsent", -1)
	source = strings.Replace(source, "\"[LogFile]\"", "con", -1)
	if !strings.ContainsRune(source, '[') {
		return source
	}

	panic("placeholder present: " + source)
}

func splitParameters(arguments string) []string {
	if !strings.ContainsRune(arguments, '"') {
		return strings.Split(arguments, " ")
	}

	panic("arguments include quotes: " + arguments)
}

var msRootCertificates = func(pem string) *x509.CertPool {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(pem)) {
		panic("pool.AppendCertsFromPEM returned false")
	}
	return pool
}(`subject=/C=US/ST=Washington/L=Redmond/O=Microsoft Corporation/CN=Microsoft Root Certificate Authority 2010
issuer=/C=US/ST=Washington/L=Redmond/O=Microsoft Corporation/CN=Microsoft Root Certificate Authority 2010
-----BEGIN CERTIFICATE-----
MIIF7TCCA9WgAwIBAgIQKMw6Jb+6RKxEmptYa0M5qjANBgkqhkiG9w0BAQsFADCB
iDELMAkGA1UEBhMCVVMxEzARBgNVBAgTCldhc2hpbmd0b24xEDAOBgNVBAcTB1Jl
ZG1vbmQxHjAcBgNVBAoTFU1pY3Jvc29mdCBDb3Jwb3JhdGlvbjEyMDAGA1UEAxMp
TWljcm9zb2Z0IFJvb3QgQ2VydGlmaWNhdGUgQXV0aG9yaXR5IDIwMTAwHhcNMTAw
NjIzMjE1NzI0WhcNMzUwNjIzMjIwNDAxWjCBiDELMAkGA1UEBhMCVVMxEzARBgNV
BAgTCldhc2hpbmd0b24xEDAOBgNVBAcTB1JlZG1vbmQxHjAcBgNVBAoTFU1pY3Jv
c29mdCBDb3Jwb3JhdGlvbjEyMDAGA1UEAxMpTWljcm9zb2Z0IFJvb3QgQ2VydGlm
aWNhdGUgQXV0aG9yaXR5IDIwMTAwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQC5CJ4o5OTsBk5QaLNBxXvrrraOr4G6IkQfZTRpTL5wQBfyFnvief2G7Q05
9BuorZKQHss9do9a2bWREC48BY2KbSRU5x/tVq2DtFCcFaUXdIhZIPwIxYR202jU
byh4zly481CQRP/jY1++oZoslhUE1gf+HoQh4EIxEcQoNpTPUKRinsnWq3EAslsM
5pbUCiSW9f/G1bcb18u3IWKvEtyhXTfjGvsaRpjAm8DnYx8qCJMCfh5qjvKfGInk
IoWisYRXQP/1DthvnO3iRTEBzRfpf7CBReOqIUAmoXKqp088AQV+7oNYsV4GY5li
kXiCtw2TDCRqtBvbJ+xflQQ/k0ow9ZcYs6f5GaeTMx0ByNsiUlzXJclG+aL7h1lD
vptisY0thkQaRqx4YX4wCfquicRBKiJmA5E5RZzHiwyoyg0v+1LqDPdjMyOd/rAf
rWfWp1ADxgRwY7UssYZaQ7f7rvluKW4hIUEmBozJw+6wwoWTobmF2eYybEtMP9Zd
o+W1nXfDnMBVt3QA47g4q4OXUOGaQiQdxsCjMNEaWshSNPdz8ccYHzOteuzLQWDz
I5QgwkhFrFxRxi6AwuJ3Fb2Fh+02nZaR7gC1o3Dsn+ONgGiDdrqvXXBSIhbiZvu6
s8XC9z4vd6bK3sGmxkhMwzdRI9Mn17hOcJbwoUR2r3jPmuFmEwIDAQABo1EwTzAL
BgNVHQ8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU1fZWy4/oolxi
aNE9lJBb186aGMQwEAYJKwYBBAGCNxUBBAMCAQAwDQYJKoZIhvcNAQELBQADggIB
AKylloy/u66m9tdxh0MxVoj9HDJxWzW31PCR8q834hTx8wImBT4WFH8UurhP+4my
sufUCcxtuVs7ZGVwZrfysVrfGgLz9VG4Z215879We+SEuSsem0CcJjT5RxiYadgc
17bRv49hwmfEte9gQ44QGzZJ5CDKrafBsSdlCfjN9Vsq0IQz8+8f8vWcC1iTN6B1
oN5y3mx1KmYi9YwGMFafQLkwqkB3FYLXi+zA07K9g8V3DB6urxlToE15cZ8PrzDO
Z/nWLMwiQXoH8pdCGM5ZeRBV3m8Q5Ljag2ZAFgloI1uXLiaaArtXjMW4umliMoCJ
nqH9wJJ8eyszGYQqY8UAaGL6n0eNmXpFOqfp7e5pQrXzgZtHVhB7/HA2hBhz6u/5
l02eMyPdJgu6Krc/RNyDJ/+9YVkrEbfKT9vFiwwcMa4y+Pi5Qvd/3GGadrFaBOER
PWZFtxhxvskkhdbz1LpBNF0SLSW5jaYTSG1LsAd9mZMJYYF0VyaKq2nj5NnHiMwk
2OxSJFwevJEU4pbe6wrant1fs1vb1ILsxiBQhyVAOvvH7s3+M+Vuw4QJVQMlOcDp
NV1lMaj2v6AJzSnHszYyLtyV84PBWs+LjfbqsyH4pO0eMQ62TBGrYAukEiMiF6M2
ZIKRBBLgq28ey1AFYbRA/1mGcdHVM2l8qXOKONdkDPFp
-----END CERTIFICATE-----
subject=/C=US/ST=Washington/L=Redmond/O=Microsoft Corporation/CN=Microsoft Root Certificate Authority 2011
issuer=/C=US/ST=Washington/L=Redmond/O=Microsoft Corporation/CN=Microsoft Root Certificate Authority 2011
-----BEGIN CERTIFICATE-----
MIIF7TCCA9WgAwIBAgIQP4vItfyfspZDtWnWbELhRDANBgkqhkiG9w0BAQsFADCB
iDELMAkGA1UEBhMCVVMxEzARBgNVBAgTCldhc2hpbmd0b24xEDAOBgNVBAcTB1Jl
ZG1vbmQxHjAcBgNVBAoTFU1pY3Jvc29mdCBDb3Jwb3JhdGlvbjEyMDAGA1UEAxMp
TWljcm9zb2Z0IFJvb3QgQ2VydGlmaWNhdGUgQXV0aG9yaXR5IDIwMTEwHhcNMTEw
MzIyMjIwNTI4WhcNMzYwMzIyMjIxMzA0WjCBiDELMAkGA1UEBhMCVVMxEzARBgNV
BAgTCldhc2hpbmd0b24xEDAOBgNVBAcTB1JlZG1vbmQxHjAcBgNVBAoTFU1pY3Jv
c29mdCBDb3Jwb3JhdGlvbjEyMDAGA1UEAxMpTWljcm9zb2Z0IFJvb3QgQ2VydGlm
aWNhdGUgQXV0aG9yaXR5IDIwMTEwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQCygEGqNThNE3IyaCJNuLLx/9VSvGzH9dJKjDbu0cJcfoyKrq8TKG/Ac+M6
ztAlqFo6be+ouFmrEyNozQwph9FvgFyPRH9dkAFSWKxRxV8qh9zc2AodwQO5e7BW
6KPeZGHCnvjzfLnsDbVU/ky2ZU+I8JxImQxCCwl8MVkXeQZ4KI2JOkwDJb5xalwL
54RgpJki49KvhKSn+9GY7Qyp3pSJ4Q6g3MDOmT3qCFK7VnnkH4S6Hri0xElcTzFL
h93dBWcmmYDgcRGjuKVB4qRTufcyKYMME782XgSzS0NHL2vikR7TmE/dQgfI6B0S
/Jmpaz6SfsjWaTr8ZL22CZ3K/QwLopt3YEsDlKQwaRLWQi3BQUzK3Kr9j1uDRprZ
/LHR47PJf0h6zSTwQY9cdNCssBAgBkm3xy0hyFfj0IbzA2j70M5xwYmZSmQBbP3s
MJHPQTySx+W6hh1hhMdfgzlirrSSL0fzC/hV66AfWdC7dJse0Hbm8ukG1xDo+mTe
acY1logC8Ea4PyeZb8txiSk190gWAjWP1Xl8TQLPX+uKg09FcYj5qQ1OcunCnAfP
SRtOBA5jUYxe2ADBVSy2xuDCZU7JNDn1nLPEfuhhbhNfFcRf2X7tHc7uROzLLoax
7Dj2cO2rXBPB2Q8Nx4CyVe0096yb5MPa50c8prWPMd/FS6/r8QIDAQABo1EwTzAL
BgNVHQ8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUci06AjGQQ7kU
BU7h6qfHMdEjiTQwEAYJKwYBBAGCNxUBBAMCAQAwDQYJKoZIhvcNAQELBQADggIB
AH9yzw+3xRXbm8BJyiZb/p4T5tPw0tuXX/JLP02zrhmu7deXoKzvqTqjwkGw5biR
nhOBJAPmCf0/V0A5ISRW0RAvS0CpNoZLtFNXmvvxfomPEf4YbFGq6O0JlbXlccmh
6Yd1phV/yX43VF50k8XDZ8wNT2uoFwxtCJJ+i92Bqi1wIcM9BhS7vyRep4TXPw8h
Ir1LAAbblxzYXtTFC1yHblCk6MM4pPvLLMWSZpuFXst6bJN8gClYW1e1QGm6CHmm
ZGIVnYeWRbVmIyADixxzoNOieTPgUFmG2y/lAiXqcyqfABTINseSO+lOAOzYVgm5
M0kS0lQLAausR7aRKX1MtHWAUgHoyoL2n8ysnI8X6i8msKtyrAv+nlEex0NVZ09R
s1fWtuzuUrc66U7h14GIvE+OdbtLqPA1qibUZ2dJsnBMO5PcHd94kIZysjik0dyS
TclY6ysSXNQ7roxrsIPlAT/4CTL2kzU0Iq/dNw13CYArzUgA8YyZGUcFAenRv9FO
0OYoQzeZpApKCNmacXPSqs0xE2N2oTdvkjgefRI8ZjLny23h/FKJ3crWZgWalmG+
oijHHKOnNlA8OqTfSm7mhzvO6/DggTedEzxSjr25HTTGHdUKaj2YKXCMiSrRq4IQ
SB/c9O+lxbtVGjhjhE63bK2VVOxlIhBJF7jAHscPrFRH
-----END CERTIFICATE-----
subject=/DC=com/DC=microsoft/CN=Microsoft Root Certificate Authority
issuer=/DC=com/DC=microsoft/CN=Microsoft Root Certificate Authority
-----BEGIN CERTIFICATE-----
MIIFmTCCA4GgAwIBAgIQea0WoUqgpa1Mc1j0BxMuZTANBgkqhkiG9w0BAQUFADBf
MRMwEQYKCZImiZPyLGQBGRYDY29tMRkwFwYKCZImiZPyLGQBGRYJbWljcm9zb2Z0
MS0wKwYDVQQDEyRNaWNyb3NvZnQgUm9vdCBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkw
HhcNMDEwNTA5MjMxOTIyWhcNMjEwNTA5MjMyODEzWjBfMRMwEQYKCZImiZPyLGQB
GRYDY29tMRkwFwYKCZImiZPyLGQBGRYJbWljcm9zb2Z0MS0wKwYDVQQDEyRNaWNy
b3NvZnQgUm9vdCBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDzXfqAZ9Rap6kMLJAg0DUIPHWEzbcHiZyJ2t7Ow2D6
kWhanpRxKRh2fMLgyCV2lA5Y+gQ0Nubfr/eAuulYCyuT5Z0F43cikfc0ZDwikR1e
4QmQvBT+/HVYGeF5tweSo66IWQjYnwfKA1j8aCltMtfSqMtL/OELSDJP5uu4rU/k
XG8TlJnbldV126gat5SRtHdb9UgMj2p5fRRwBH1tr5D12nDYR7e/my9s5wW34RFg
rHmRFHzF1qbk4X7Vw37lktI8ALU2gt554W3ztW74nzPJy1J9c5g224uha6KVl5uj
3sJNJv8GlmclBsjnrOTuEjOVMZnINQhONMp5U9W1vmMyWUA2wKVOBE0921sHM+RY
v+8/U2TYQlk1V/0PRXwkBE2e1jh0EZcikM5oRHSSb9VLb7CG48c2QqDQ/MHAWvmj
YbkwR3GWChawkcBCle8Qfyhq4yofseTNAz93cQTHIPxJDx1FiKTXy36IrY4t7EXb
xFEEySr87IaemhGXW97OU4jm4rf9rJXCKEDb7wSQ34EzOdmyRaUjhwalVYkxuwYt
YA5BGH0fLrWXyxHrFdUkpZTvFRSJ/Utz+jJb/NEzAPlZYnAHMuouq0Ate8rdIWcb
MJmPFqojqEHRsG4RmzbE3kB0nOFYZcFgHnpbOMiPuwQmfNQWQOW2a2yqhv0Av87B
NQIDAQABo1EwTzALBgNVHQ8EBAMCAcYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4E
FgQUDqyCYEBWJ5flJRP8KuEKU5VZ5KQwEAYJKwYBBAGCNxUBBAMCAQAwDQYJKoZI
hvcNAQEFBQADggIBAMURTQM6YN1dUhF3j7K7NsiyBb+0t6jYIJ1cEwO2HCL6BhM1
tshj1JpHbyZX0lXxBLEmX9apUGigvNK4bszD6azfGc14rFl0rGY0NsQbPmw4TDMO
MBINoyb+UVMA/69aToQNDx/kbQUuToVLjWwzb1TSZKu/UK99ejmgN+1jAw/8EwbO
FjbUVDuVG1FiOuVNF9QFOZKaJ6hbqr3su77jIIlgcWxWs6UT0G0OI36VA+1oPfLY
Y7hrTbboMLXhypRL96KqXZkwsj2nwlFsKCABJCcrSwC3nRFrcL6yEIK8DJto0I07
JIeqmShynTNfWZC99d6TnjpiWjQ54ohVHbkGsMGJay3XacMZEjaE0Mmg2v8vaXiy
5Xra69cMwPe9Yxe4ORM4ojZbe/KFVmodZGLBOOKqv1FmopT1EpxmIhBr8rcwki3y
KfA9OxRDaKLxnCk3y844ICVtfGfzfiQSJAMIgUfspZ6X9RjXz7vV73aW7/3O21ad
laBC+ZdY4dcxItNfWeY+biIA6kOEtiXb2fMIVmjAZGsdfOy2k6JiV24u2OdYj8Qx
SSbd3ik1h/UwcXBbFDxpvYkSfesuo/7Yf56CWlIKK8FDK9kwiJ/IEPuJjeahhXUz
fmye23MTZGJppS99ypZtn/gETTCSPW4hFCHJPeDD/YprnUr90aGdmUN3P7Da
-----END CERTIFICATE-----
`)

func validateSignature(data []byte) {
	log.Println("TODO: validate signature")
	return

	newLine := bytes.IndexByte(data, '\n')
	if data[newLine-1] == '\r' && data[newLine-2] == ',' {
		data[newLine-2] = '}'
	} else {
		panic("unexpected signed manifest format")
	}
	newLine++

	payload := data[:newLine]
	log.Println(string(payload))
	var signature struct {
		Signature struct {
			SignInfo struct {
				SignatureMethod  string `json:"signatureMethod"`
				DigestMethod     string `json:"digestMethod"`
				DigestValue      []byte `json:"digestValue"`
				Canonicalization string `json:"canonicalization"`
			} `json:"signInfo"`
			SignatureValue []byte `json:"signatureValue"`
			KeyInfo        struct {
				KeyValue struct {
					RSAKeyValue struct {
						Modulus  []byte `json:"modulus"`
						Exponent []byte `json:"exponent"`
					} `json:"rsaKeyValue"`
				} `json:"keyValue"`
				X509Data [1][]byte `json:"x509Data"`
			} `json:"keyInfo"`
		} `json:"signature"`
	}
	err := json.Unmarshal(append([]byte{'{'}, data[newLine:]...), &signature)
	if err != nil {
		panic(err)
	}

	spew.Dump(sha1.Sum(payload))
	spew.Dump(signature)

	panic("TODO")
}

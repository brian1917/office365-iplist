package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/brian1917/illumioapi"
)

type msendpoint struct {
	ID                     int      `json:"id,omitempty"`
	ServiceArea            string   `json:"serviceArea,omitempty"`
	ServiceAreaDisplayName string   `json:"serviceAreaDisplayName,omitempty"`
	Urls                   []string `json:"urls,omitempty"`
	Ips                    []string `json:"ips,omitempty"`
	TCPPorts               string   `json:"tcpPorts,omitempty"`
	ExpressRoute           bool     `json:"expressRoute,omitempty"`
	Category               string   `json:"category,omitempty"`
	Required               bool     `json:"required,omitempty"`
	Notes                  string   `json:"notes,omitempty"`
	UDPPorts               string   `json:"udpPorts,omitempty"`
}

func generateGUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func main() {

	fqdn := flag.String("fqdn", "", "The fully qualified domain name of the PCE.")
	port := flag.Int("port", 8443, "The port for the PCE.")
	user := flag.String("user", "", "API user or email address.")
	pwd := flag.String("pwd", "", "API key if using API user or password if using email address.")
	disableTLS := flag.Bool("x", false, "Disable TLS checking.")
	provision := flag.Bool("p", false, "Provision the IP List.")
	name := flag.String("name", "Office365", "Name of the IPList to be created or updated")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("-fqdn  string")
		fmt.Println("       The fully qualified domain name of the PCE. Required.")
		fmt.Println("-port  int")
		fmt.Println("       The port of the PCE. (default 8443)")
		fmt.Println("-user  string")
		fmt.Println("       API user or email address. Required.")
		fmt.Println("-pwd   string")
		fmt.Println("       API key if using API user or password if using email address. Required.")
		fmt.Println("-name  string")
		fmt.Println("       Name of the IPList to be created or updated. (default Office365)")
		fmt.Println("-p     Provision the IP List.")
		fmt.Println("-x     Disable TLS checking.")

	}

	// Parse flags
	flag.Parse()

	// Run some checks on the required fields
	if len(*fqdn) == 0 || len(*user) == 0 || len(*pwd) == 0 {
		log.Fatalf("ERROR - Required arguments not included. Run -h for usgae.")
	}

	// Build the PCE from user input
	pce, err := illumioapi.PCEbuilder(*fqdn, *user, *pwd, *port, *disableTLS)
	if err != nil {
		log.Fatalf("Error building PCE - %s", err)
	}

	// Get the IP information from Microsoft
	var msendpoints []msendpoint

	resp, err := http.Get("https://endpoints.office.com/endpoints/worldwide?clientrequestid=" + generateGUID())
	if err != nil {
		log.Fatalf("Error calling Microsoft web service - %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading Microsoft response - %s", err)
	}

	err = json.Unmarshal(body, &msendpoints)
	if err != nil {
		log.Fatalf("Error unmarshaling Microsoft response - %s", err)
	}

	// Update or create the IP List
	var ipRanges []*illumioapi.IPRange
	uniqueIPs := make(map[string]int)
	for _, mse := range msendpoints {
		for _, ip := range mse.Ips {
			if _, ok := uniqueIPs[ip]; !ok {
				ipRanges = append(ipRanges, &illumioapi.IPRange{FromIP: ip})
				uniqueIPs[ip] = 1
			}

		}
	}
	ipList := illumioapi.IPList{Name: *name, IPRanges: ipRanges}

	// Get all IP Lists
	allDraftIPL, _, err := illumioapi.GetAllIPLists(pce, "draft")
	if err != nil {
		log.Fatalf("Error getting draft IP Lists - %s", err)
	}
	allActiveIPL, _, err := illumioapi.GetAllIPLists(pce, "active")
	if err != nil {
		log.Fatalf("Error getting active IP Lists - %s", err)
	}
	allIPL := append(allActiveIPL, allDraftIPL...)

	// Get the href if there is an existing IP List that matches our name
	var href string
	for _, ipl := range allIPL {
		if ipl.Name == *name {
			href = ipl.Href
		}
	}

	// If the href is blank, we need to create the IP List
	if href == "" {
		ipl, _, err := illumioapi.CreateIPList(pce, ipList)
		if err != nil {
			log.Fatalf("Error creating IP List - %s", err)
		}
		fmt.Println("IP List created")
		if *provision {
			_, err = illumioapi.ProvisionHref(pce, ipl.Href)
			if err != nil {
				log.Fatalf("Error provisioning IP List - %s", err)
			}
			fmt.Println("IP List Provisioned")
		}
	} else {
		// If the href is not blank, we have to update the IP List
		ipList.Href = href
		_, err = illumioapi.UpdateIPList(pce, ipList)
		if err != nil {
			log.Fatalf("Error updating IP list - %s", err)
		}
		fmt.Println("IP List updated")
		if *provision {
			_, err = illumioapi.ProvisionHref(pce, href)
			if err != nil {
				log.Fatalf("Error provisioning IP List - %s", err)
			}
			fmt.Println("IP List Provisioned")
		}
	}

}

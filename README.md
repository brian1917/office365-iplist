# Office 365 IP List Creater
Binaries for Mac, Linux, and Windows are the `bin` folder of this repository.

## Description
CLI tool to run on a cron job to keep an Office 365 Illumio IP List up to date based on the Microsoft web service documented here: https://docs.microsoft.com/en-us/office365/enterprise/urls-and-ip-address-ranges.
On the first run the tool will create the IP List and on subsequent runs it will update it (based on `name` parameter).

## Usage
`office365-iplist -h`

```
Usage of office365-iplist:
-fqdn  string
       The fully qualified domain name of the PCE. Required.
-port  int
       The port of the PCE. (default 8443)
-user  string
       API user or email address. Required.
-pwd   string
       API key if using API user or password if using email address. Required.
-name  string
       Name of the IPList to be created or updated. (default Office365)
-p     Provision the IP List.
-x     Disable TLS checking.
```

# DNSCheck

## Build
For amd64: `CGO_ENABLED=1 CC="zig cc -target native-native-musl" CXX="zig cc -target native-native-musl" go build`
Form arm: `CGO_ENABLED=1 GOOS=linux GOARCH=arm CC="zig cc -target arm-linux-musleabihf" CXX="zig c++ -target arm-linux-musleabihf" go build`

## Prepare Database

Once the server is started the first time, it creates the table `ProcessedDomains` to store the domains that were already checked including the used variations.

To fill this table, open the SQLite file with e.g. SQLiteBrowser and run the following Query:

```sql
INSERT OR IGNORE INTO tblProcessedDomains 

SELECT Domain, NULL, NULL,NULL, NULL from
(SELECT DISTINCT LOWER(DOMAIN) AS Domain from tblAvailableDomains 
UNION
SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomainDNSData 
UNION
SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomains)
```
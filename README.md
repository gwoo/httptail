# Httptail
Have fun streaming logs over http. Provide insight into system services without needing to ssh into a server.


Uses server sent events to keep the connection alive and stream the output of the log file.


In a browser, access "http://host:port/file.log" and wait for each line to be printed as it is written to the log.

	Usage of ./httptail:
	  -addr=":2222": Addr for server
	  -creds="admin:password": Authentication credentials.
	  -dir="/var/log": Base directory where logs are found.


## Adding SSL

Files are served through SSL, if a `cert.pem` and `key.pem` exist in the current working directory where `httptail` is started.

To generate a cert and key, use openssl

	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem

## License
The BSD License http://opensource.org/licenses/bsd-license.php.
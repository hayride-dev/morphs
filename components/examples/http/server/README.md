## Hayride HTTP Server Example

This example demonstrates how to create a simple HTTP server using Hayride. The server listens for incoming HTTP requests and responds with a JSON object containing the request method and URL.

### Casting 

This example uses the options flag when casting a morph. 
This allows you to specify the address and port for the server to listen on. 

`hayride cast --morph hayride-examples:server --options address=localhost:8088` 

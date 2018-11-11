# Example usage
The supplied docker-compose file contains a registry:2 and a mikaellindemann/registryfrontend service.

The environment of the frontend is set to add the registry on startup, using the internal docker network.
Furthermore, the example will expose the frontend on port 80 on the host.

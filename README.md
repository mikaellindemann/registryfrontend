# A docker registry v2 frontend written in Go

There are many other frontends for the docker registry out there. The best one I have found yet, is [docker-registry-frontend](https://github.com/brennerm/docker-registry-frontend) by [Max Brenner](https://github.com/brennerm).

The only downside to the above solution, is that the provided image is very large, and that it runs single-threaded in the Flask development server.

Therefore I decided to make a simple docker registry v2 frontend myself, greatly inspired by the above service.
While my version does not offer all of the same features, it has the benefit of taking up very little space, while still providing a service that can be used by multiple viewers at the same time.

The source can be found at GitHub: [https://github.com/mikaellindemann/registryfrontend](https://github.com/mikaellindemann/registryfrontend).

The images can be pulled from Docker Hub: [https://hub.docker.com/r/mikaellindemann/registryfrontend](https://hub.docker.com/r/mikaellindemann/registryfrontend).

## Features
Currently, the frontend supports multiple docker registry v2 instances, that are publicly available or protected by Basic authentication.

One registry can be added on startup by using the following environment variables:

| Name | Description |
| ---- | ----------- |
| REGISTRY_NAME | The identifier used in the URL when viewing information about the registry and it's content. |
| REGISTRY_URL  | The URL poiting to the registry. |
| REGISTRY_AUTH_BASIC_USER | The Basic authentication username. |
| REGISTRY_AUTH_BASIC_PASSWORD | The Basic authentication password. |

Note that the registry will only be added if the name contains only legal characters (a-z and 0-9) and if both a name and URL is provided.

Registries added through the frontend, will currently not be persisted on restart of the frontend.

## Known issues/bugs
* If multiple users are adding or removing registries in the same instance, the service might panic due to no serialized access to the in-memory storage.
* If a registry has namespaced repositories, the frontend will treat the namespace as a repository, and the repository as a tag.

Pull requests and issues are very welcome.

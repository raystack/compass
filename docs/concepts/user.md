# User
The current version of Compass does not have user management. Compass expects there is an external instance that manages user. Compass consumes user information from the configurable identity uuid header in every API call. The default name of the header is `Compass-User-UUID`. 
Compass does not make any assumption of what kind of identity format that is being used. The `uuid` indicates that it could be in any form (e.g. email, UUIDv4, etc) as long as it is universally unique.
The current behaviour is, Compass will add a new user if the user information consumed from the header does not exist in Compass' database. 

## Phantom User
In Compass ingestion API, Compass allows asset to mentioned who is its own owners. During the ingestion, if the `email` field in the list of `owners` field in the asset is not empty, Compass will create a new `'Phantom User'` with the email but with empty UUID.
A `'Phantom User'` is a user that is written in the storage but with empty UUID. The `'Phantom User'` cannot do any user-related interaction (e.g. Starring, Discussion) in Compass.

## Linking User
There is another configurable optional email header that Compass expect. The default name is `Compass-User-Email`. In case there is already an existing `'Phantom User'`, if there is a request coming to Compass with completed user information in its header (uuid header and email header are not empty), Compass will register the UUID to the existing `'Phantom User'` and the `'Phantom User'` becomes a normal user. By doing so, assets ownership of that new user will immediately reflected.

## User Provider
Since Compass expects that there is an external instance that manages user, it is possible for Compass to consume user information from multiple external instances. Compass distinguishes the source of user by marking it in the `provider` field. The default `provider` field value can be configured via config.
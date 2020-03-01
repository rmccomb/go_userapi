# go_userapi
A secured REST interface to a user manager written in go

Methods available are:

GET /status 
    Check the seb server status

POST /signin
    A valid username(email) & password should be provided in the request body in JSON
    The method returns a JWT token which is used to authenticate in the other methods.

GET /validatetoken 
    Validate the given token and return OK on success

GET /users
    Get all users in the backend store

GET /user/{email}
    Get a specific user

PUT /user
    Add (register) a new user to the store (non secured)

POST /user
    Update an existing user in the store

DELETE /user/{email}
    Delete a specific user in the store

The back end store is an in-memory cache for simplicity.

Configuration is in the config.yaml file. 
Elements are:
AdminEmail and AdminPassword for the admin user/
JWTSigningKey is used to sign the token

A front-end is available in another repository
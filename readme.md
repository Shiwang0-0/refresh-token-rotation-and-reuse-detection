
# Behaviour

- if refresh token expiration is 10 second and user log in and refreshes after 10 second the refresh token is deleted from the cookie and the database as well
- if refresh token expiration is 10 seconds and user log in and log out before 10 second the refresh token is deleted from the cookie and the database as well

- if refresh token expiration is 30 second and access token expiration is 10 second, we will get new access and new refresh token in this duration. every refresh token ROTATES on every access token expiration.

- if refresh token is expired, and we use it after expiration (by replacing the current refresh token), we will be redirected to the login page again

- while for `/refresh` route when token is found to be expired a database query is fired to set `refreshToken=NULL` but when the token has time left before expiration but found to be invalid for some reason (signature mismatch, wrong secret, not valid jwt format, wrong siging algorithm, missing claims), then it still sits on the database. probably a async job should be clearing these types of refreshTokens (out of scope currently)
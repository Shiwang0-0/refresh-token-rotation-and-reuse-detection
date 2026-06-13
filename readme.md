
# Referesh Tokens



# Behaviour

#### refresh tokens
- if refresh token expiration is 10 second and user log in and refreshes after 10 second the refresh token is deleted from the cookie and the database as well
- if refresh token expiration is 10 seconds and user log in and log out before 10 second the refresh token is deleted from the cookie and the database as well

- if refresh token expiration is 30 second and access token expiration is 10 second, we will get new access and new refresh token in this duration. every refresh token ROTATES on every access token expiration.

- if refresh token is expired, and we use it after expiration (by replacing the current refresh token), we will be redirected to the login page again

- if user login and closes the browser and comes back after session expiration, then a new login will be required. but for the previous refresh token stored it will still be in the database, similary if user looses the refresh tokens the access to the session is lost therefore the database row for the session still persist.
for these kind of ideal rows where `expirationTime < time.Now()` we need a periodic job (out of scope for current implementation)

#### reuse detection
- if a valid user log in and the hacker stole the refresh token, and at the same time valid user's token gets rotated (removed from DB) then in the hacker's session there exist a session where refresh token is present in the hacker browser request cookie but no user has that token in the DB because of the rotation from the valid user's side (meaning a reuse occured).To prevent this reuse, all the sessions belonging to that user are deleted from the database. Therefore no matter if previous sessions were stolen, hacker cannot login



# some frontend flow decisions
`/refresh` is only called when accessToken is null which only happens on first load or hard page refresh. 

- Navigate to `/profile`(first load) 
- accessToken = null in context
- `/refresh` called

- Navigate to `/profile`
- accessToken exists in context
- return early, skip `/refresh`

#### ProtectedRoute /refresh and api /refresh
The accessToken check in ProtectedRoute only runs on navigation, not during a session:
User logs in and access token is stored in context (15 min lifetime)

1) can we just use a protected route ??  
    User sits on `/profile` for 20 minutes doing stuff
    - no navigation happens
    - ProtectedRoute never remounts
    - access token silently expires in context

    User clicks a button
    - api("/dashboard/data") called (hypothetical)
    - access token is expired
    - 401 received <<<--- this is where api.js 401 comes
    - /refresh called
    - retry

    So they handle two different scenarios:  
    - ProtectedRoute /refresh check
        - handles: page load, hard refresh
        - when: on navigation, context is empty

    - api.js 401 handling
        - handles: token expiring mid-session while user is active on a page
        - when: user is already on a page, token expires, they make an API call

    Without api.js 401 handling:
    User on /dashboard for 20 min
    - token expires
    - clicks button and api call fails with 401
    - nothing handles it


2) cant we just set access token by 401 error ??  

    User hard refreshes /dashboard
    - context is wiped and accessToken = null
    - ProtectedRoute renders children immediately
    - Home component mounts
    - api("/dashboard/data") called with null token
    - backend receives request with no token
    - backend returns 401
    - api.js calls /refresh and gets new token
    - retries request

    it works, but the problem is:
    context = null
    - children render for a split second with no user data
    - UI flashes / breaks because user is null
    - then data loads


ProtectedRoute: guards the route, populates user before render
api.js 401 handling: handles token expiring while user is active
Both are needed for different reasons. ProtectedRoute is about initial load safety, api.js is about mid-session recovery.



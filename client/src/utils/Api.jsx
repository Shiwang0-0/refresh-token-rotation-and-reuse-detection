import { useAuth } from "../context/useAuth"
import { BASE_URL } from "../App"
import { useEffect, useRef } from "react"

// this hits the refresh end point when on the first try the auth middleware of backend returns unauthorized becuase the accessToken that was sent expired
// and unlinke the /refresh hit of the useEffect in ProtectedRoute
// this hits even if the page is not changed (the ProtectedRoute does not re-run) and on the same page the access token expires
const refreshTokens = async (setAccessToken) => {
    const response = await fetch(`${BASE_URL}/refresh`, {
        method: "POST",
        credentials: "include", // sends refresh_token cookie
    })

    if (!response.ok) {
        window.location.href = "/login"
        return null
    }

    const data = await response.json()
    // backend already set new refresh_token cookie (rotation)
    setAccessToken(data.access_token) // update module-level token
    return data.access_token
}

export function useApi() {
    const { accessToken, setAccessToken } = useAuth();

    // the context might use old value, to actually get new rotated access token do this
    const accessTokenRef = useRef(accessToken);
    useEffect(() => {
        accessTokenRef.current = accessToken;
    }, [accessToken]);

    useEffect(() => {
        console.log("ACCESS TOKEN CHANGED:", accessToken);
    }, [accessToken]);

    const api = async (url, options = {}) => {

        // a normal call to the request api is made, with 
        let response = await fetch(`${BASE_URL}${url}`, {
            ...options,
            credentials: "include",
            headers: {
                "Content-Type": "application/json",
                ...(accessTokenRef.current && {
                    Authorization: `Bearer ${accessTokenRef.current}`, // read from the reference
                }), // sends the access token manually, not in a cookie, this prevents CSRF by not attaching the accessToken directly in the cookie
                ...options.headers,
            },
        })

        console.log("TOKEN SENT:", accessTokenRef.current);

        // if there is a response of unauthorized from the auth middleware of backend
        // that means the access token is no more valid (the auth middleware validates the access token sent in the header)
        if (response.status === 401) {
            try {
                const newToken = await refreshTokens(setAccessToken); // rotates both tokens
                if (!newToken) return;
                // retry original request with new access token (and rotated refresh token too)
                response = await fetch(`${BASE_URL}${url}`, {
                    ...options,
                    credentials: "include",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${newToken}`,
                        ...options.headers,
                    },
                });
            } catch (err) {
                setAccessToken(null);
                window.location.href = "/login";
                throw err;
            }
        }

        if (!response.ok) {
            const data = await response.json().catch(() => ({}));

            throw new Error(
                data.message || `Request failed (${response.status})`
            );
        }

        return response
    }
    return api;
}
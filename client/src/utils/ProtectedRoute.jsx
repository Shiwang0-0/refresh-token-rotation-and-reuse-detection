import { useEffect, useState } from 'react';
import { useAuth } from '../context/useAuth'
import { Navigate, Outlet } from 'react-router-dom';
import { BASE_URL } from '../App';

// this protected route, checks if the access token exist or not
// if it doesnt then it makes a refresh call to the backend, which returns the access token and also rotate the session token in the DB internally

// if access token exist, there is no need to generate a new one or even change the exising session token (session rotation)
// session rotation only happens when the access token expires


export const ProtectedRoute = () => {

    const {accessToken, setAccessToken, setUser} = useAuth()

    const [loading, setLoading] = useState(!accessToken);       // false if token exists
    const [isAuthenticated, setIsAuthenticated] = useState(!!accessToken); // true if token exists

    useEffect(()=>{

        const initializeAuth = async ()=>{

            // the access token is already present for the protected route
            // it is not expired yet
            if (accessToken) {
                setIsAuthenticated(true);
                setLoading(false);
                return;
            }

            console.log("dont have access token")
            
            // either it is the first time (hard refresh) or the access token expired
            try {
                // these doesnt need the api wrapper
                const refreshResponse = await fetch(`${BASE_URL}/refresh`, {
                    method: "POST",
                    credentials: "include",
                });

                const refreshData = await refreshResponse.json();

                if (!refreshResponse.ok) {
                    throw new Error(refreshData.message || "Refresh failed");
                }

                setAccessToken(refreshData.access_token);

                // these doesnt need the api wrapper
                const profileResponse = await fetch(`${BASE_URL}/profile`, {
                    method: "GET",
                    headers: {
                        Authorization: `Bearer ${refreshData.access_token}`,
                    },
                    credentials: "include",
                });

                const profileData = await profileResponse.json();

                if (!profileResponse.ok) {
                    throw new Error(profileData.message || "Failed to fetch profile");
                }
                console.log("user profile: ", profileData.user)
                setUser(profileData.user);
                setIsAuthenticated(true);
            } catch (err) {
                console.error(err);
                setIsAuthenticated(false);
            } finally {
                setLoading(false);
            }
        }

        initializeAuth();
    },[accessToken, setAccessToken, setUser])

    if (loading) return <div>Loading...</div>;
    
    return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />;
}
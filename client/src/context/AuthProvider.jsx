import { useState } from "react";
import { AuthContext } from "./AuthContext";

export function AuthProvider({ children }) {
    const [accessToken, setAccessToken] = useState(null);
    const [user, setUser] = useState(null);

    return (
        <AuthContext.Provider
            value={{ accessToken, setAccessToken, user, setUser }}
        >
            {children}
        </AuthContext.Provider>
    );
}
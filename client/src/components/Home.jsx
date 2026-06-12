import { useAuth } from "../context/useAuth";
import { useNavigate } from "react-router-dom";
import { BASE_URL } from "../App";
import { useState } from "react";
import { useApi } from "../utils/Api";

const Home = () => {
    const { user, setUser, setAccessToken } = useAuth();
    const [profileData, setProfileData] = useState(null);
    const api = useApi()
    const navigate = useNavigate();

    // dont need the api wrapper, since we dont need the access token to send
    // delete the refresh token from cookie and database to logout
    const handleLogout = async () => {
    try {
            await fetch(`${BASE_URL}/logout`, {
                method: "POST",
                credentials: "include",
            });
        } catch (err) {
            console.error(err);
        } finally {
            setUser(null);
            setAccessToken(null);
            navigate("/login");
        }
    };

    const handleFetchProfile = async () => {
        try {
            const response = await api("/profile");
            const data = await response.json();
            console.log("profile data:", data);
            setProfileData(data.user);
        } catch (err) {
            console.error(err);
        }
    };

    if (!user) return <div>Loading user...</div>;

    return (
        <div className="flex flex-col items-center justify-center h-screen gap-6">
            <h1 className="text-4xl">
                Welcome {user.name} - {user.email}
            </h1>
            <button
                onClick={handleFetchProfile}
                className="bg-indigo-500 hover:bg-indigo-600 transition-all text-white px-6 py-2 rounded-md cursor-pointer"
            >
                Fetch Profile (test rotation)
            </button>
             {profileData && (
                <div className="text-sm text-gray-500">
                    <p>name: {profileData.name}</p>
                    <p>email: {profileData.email}</p>
                </div>
            )}
            <button
                onClick={handleLogout}
                className="bg-red-500 hover:bg-red-600 transition-all text-white px-6 py-2 rounded-md cursor-pointer"
            >
                Logout
            </button>
        </div>
    );
};

export default Home;
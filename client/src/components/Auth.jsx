import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/useAuth";
import { BASE_URL } from "../App";

const Auth = () => {
    const navigate = useNavigate()
    const [state, setState] = useState("login");
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("")
    const [loading, setLoading] = useState(false);

    const { setUser, setAccessToken } = useAuth();

    const handleAuth=async (e)=>{
        e.preventDefault()
        setError("")
        setLoading(true)
        const user = state === "login" ? { email, password } : { name, email, password };
        console.log("user:",user)

        const endpoint = (state === "login" ? "/login" : "/register")

        try{

        // no need api wrapper, this is a public route
        const response = await fetch(`${BASE_URL}${endpoint}`, {
            method: "POST",
            credentials: "include",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(user)
        })

        const data = await response.json();
        if(!response.ok){
            throw new Error(data.message || "Something went wrong")
        }
        setAccessToken(data.access_token)
        setUser(data.user)
        console.log("login and register:",data.user)
        navigate("/")
       }catch(err){
            console.error(err)
            setError(err.message)
       }finally{
        setLoading(false)
       }
    }

    if (loading){
        return <div>Loading...</div>;
    } 

    return (
        <form onSubmit={handleAuth} className="flex flex-col gap-4 m-auto items-start p-8 py-12 w-80 sm:w-[352px] text-gray-500 rounded-lg shadow-xl border border-gray-200 bg-white" >
            <p className="text-2xl font-medium m-auto">
                <span className="text-indigo-500">User</span> {state === "login" ? "Login" : "Sign Up"}
            </p>
            {error && (
                <p className="text-red-500 text-sm w-full text-center bg-red-50 p-2 rounded">
                    {error}
                </p>
            )}
            {state === "register" && (
                <div className="w-full">
                    <p>Name</p>
                    <input onChange={(e) => setName(e.target.value)} value={name} placeholder="type here" className="border border-gray-200 rounded w-full p-2 mt-1 outline-indigo-500" type="text" required />
                </div>
            )}
            <div className="w-full ">
                <p>Email</p>
                <input onChange={(e) => setEmail(e.target.value)} value={email} placeholder="type here" className="border border-gray-200 rounded w-full p-2 mt-1 outline-indigo-500" type="email" required />
            </div>
            <div className="w-full ">
                <p>Password</p>
                <input onChange={(e) => setPassword(e.target.value)} value={password} placeholder="type here" className="border border-gray-200 rounded w-full p-2 mt-1 outline-indigo-500" type="password" required />
            </div>
            {state === "register" ? (
                <p>
                    Already have account? <span onClick={() => setState("login")} className="text-indigo-500 cursor-pointer">click here</span>
                </p>
            ) : (
                <p>
                    Create an account? <span onClick={() => setState("register")} className="text-indigo-500 cursor-pointer">click here</span>
                </p>
            )}
            <button className="bg-indigo-500 hover:bg-indigo-600 transition-all text-white w-full py-2 rounded-md cursor-pointer">
                {state === "register" ? "Create Account" : "Login"}
            </button>
        </form>
    );
};

export default Auth
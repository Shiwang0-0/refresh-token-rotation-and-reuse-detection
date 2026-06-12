import Home from "./components/Home"
import Auth from "./components/Auth"
import { BrowserRouter, Routes, Route } from "react-router-dom"
import {ProtectedRoute} from "./utils/ProtectedRoute.jsx"

export const BASE_URL = "http://localhost:8000/api"

const App = () => {
  return (
        <BrowserRouter>
            <Routes>
                <Route path="/login" element={<Auth />} />

                <Route element={<ProtectedRoute />}>
                    <Route path="/" element={<Home />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}

export default App
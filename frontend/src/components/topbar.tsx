import React, { useEffect, useRef, useState } from "react";
import { ChevronDown, LogOut } from "lucide-react";
import logo from "../assets/logo.png";
import { AnimatePresence, motion } from "framer-motion";
import axios from "axios";
const Topbar: React.FC = () => {
  const [open, setOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [user, setUser] = useState<{ username: string; role: string } | null>(
    null
  );

  useEffect(() => {
    axios
      .get("http://localhost:8080/api/me", { withCredentials: true }) // penting: kirim cookie
      .then((res) => setUser(res.data))
      .catch(() => setUser({ username: "Guest", role: "User" }));
  }, []);
  // Tutup dropdown kalau klik di luar area
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const handleLogout = () => {
    localStorage.removeItem("token");
    window.location.href = "/login";
  };

  return (
    <header className="fixed top-0 left-64 right-2 h-16 flex justify-between items-center bg-white shadow z-20 rounded-xl ml-5 mt-2">
      <div className="justify-items-star felx items-center">
        <img src={logo} alt="" className="w-20 ml-2" />
      </div>
      <div className="flex items-center  justify-items-start">
        <span className="font-semibold text-black">
          {user?.username.toUpperCase() || "User"}
        </span>
        <div className="relative inline-block text-left" ref={dropdownRef}>
          <button
            onClick={() => setOpen(!open)}
            className="inline-flex items-center px-2 py-2 mr-2 bg-transparent text-black rounded-xl  transition hover:outline-none hover:ring-0 focus:outline-none focus:ring-0"
          >
            <ChevronDown className=" w-4 h-4" />
          </button>

          <AnimatePresence>
            {open && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.15 }}
                className="absolute right-0 mt-4 w-36 bg-white rounded-xl shadow-lg ring-1 ring-black/5 z-20"
              >
                <button
                  onClick={handleLogout}
                  className="flex items-center w-full px-4 py-5 rounded-xl text-sm font-medium text-gray-500 bg-transparent hover:bg-red-500 hover:text-black transition"
                >
                  <LogOut className="w-4 h-4" />
                  Logout
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </header>
  );
};
export default Topbar;

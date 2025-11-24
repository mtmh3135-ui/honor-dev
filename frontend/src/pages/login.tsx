/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
import { useState } from "react";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import logo from "../assets/logo.png";
import section from "../assets/leftdashboard.png";
import { Eye, EyeOff } from "lucide-react";
import Swal from "sweetalert2";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const navigate = useNavigate();

  const submit = async (e: any) => {
    e.preventDefault();
    try {
      const res = await axios.post(
        "http://localhost:8080/api/login",
        {
          username,
          password,
        },
        { withCredentials: true }
      );

      if (res.status === 200) {
        Swal.fire({
          title: "Login Berhasil",
          text: `Selamat datang kembali, ${username}`,
          icon: "success",
          showConfirmButton: false,
          timer: 1500,
          width: "360px",
          customClass: {
            popup: "rounded-2xl shadow-lg p-4", // padding kecil biar gak terlalu luas
            title: "text-lg font-semibold text-gray-600",
            htmlContainer: "text-sm text-gray-600",
            confirmButton:
              "bg-green-500 hover:bg-green-600 text-white text-sm rounded-lg px-4 py-1.5",
          },
        });
      }

      setTimeout(() => {
        navigate("/dashboard");
      }, 1500);
    } catch (err: any) {
      Swal.fire({
        title: "Login Gagal",
        text:
          err.response?.data?.message ||
          "Periksa kembali username dan password Anda.",
        icon: "error",
        confirmButtonColor: "#3085d6",
        width: "360px",
        customClass: {
          popup: "rounded-2xl shadow-lg p-4", // padding kecil biar gak terlalu luas
          title: "text-lg font-semibold text-red-400",
          htmlContainer: "text-sm text-gray-600",
          confirmButton:
            "bg-gray-500 hover:bg-gray-600 text-white text-sm rounded-lg px-4 py-1.5",
        },
      });
    }
  };
  return (
    <div className="flex h-screen">
      {/* Bagian Kiri (70%) */}
      <div className="w-7/12 h-screen flex flex-col items-center justify-center ">
        <img src={section} alt="Ilustrasi" className="w-3/5 " />
        <p className="text-xs text-gray-400 mt-6">
          © 2025 Honor App All rights reserved.
        </p>
      </div>

      {/* Bagian Kanan (30%) */}
      <div className="w-5/12 h-screen flex items-center justify-center">
        <div className="w-full max-w-[320px]">
          <img src={logo} alt="Logo" className="h-16 mb-6" />
          <h2 className="text-2xl font-semibold text-center mb-2 text-gray-500 ">
            Selamat Datang di <span className="text-[#92E3A9]">Honor App!</span>{" "}
          </h2>
          <p className="text-sm text-gray-500 text-center mb-6">
            Silahkan Login Akun Anda
          </p>

          <form className="space-y-4" onSubmit={submit}>
            <div>
              <h2 className="text-gray-600">Username</h2>
              <input
                type="username"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-transparent p-2 border rounded focus:ring-2 focus:ring-green-400 text-gray-500 outline-none"
              />
            </div>
            <div className="relative">
              <h2 className="text-gray-600">Password</h2>
              <div
                className="flex items-center border rounded text-gray-500 
                  focus-within:ring-2 focus-within:ring-green-400"
              >
                <input
                  type={showPassword ? "text" : "password"}
                  placeholder="••••••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-3/4 bg-transparent p-2  text-gray-500 outline-none"
                />

                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="bg-transparent text-gray-500 hover:text-green-400 
                 focus:outline-none focus:ring-0 focus:border-none 
                 active:outline-none border-none transition transform hover:scale-110 absolute right-0 bottom-0.3  "
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>

            <button
              className="w-full bg-green-500 text-white py-2 rounded-lg hover:bg-green-600 transition focus:outline-none"
              type="submit"
            >
              Login
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}

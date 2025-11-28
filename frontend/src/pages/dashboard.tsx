/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";
import axios from "axios";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { motion } from "framer-motion";
import { Typewriter } from "react-simple-typewriter";
import Swal from "sweetalert2";
import { ChevronDown, ChevronLeft, ChevronRight, Search } from "lucide-react";
import React from "react";
interface HonorDokter {
  DoctorName: string;
  CareproviderTxnDoctorId: number;
  TotalHonor: number;
}
function FuturisticHeader() {
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const handleMouseMove = (e: {
    currentTarget: { getBoundingClientRect: () => any };
    clientX: number;
    clientY: number;
  }) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width - 0.5) * 10;
    const y = ((e.clientY - rect.top) / rect.height - 0.5) * 10;

    setOffset({ x, y });
  };
  return (
    <motion.div
      onMouseMove={handleMouseMove}
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      className="relative w-full rounded-xl py-10 mb-8 bg-white border-b border-gray-200 shadow-sm overflow-hidden"
      style={{
        perspective: "1000px",
      }}
    >
      <div className="relative max-w-6xl mx-auto px-6 text-center">
        {/* Title */}
        <motion.h1
          initial={{ opacity: 0, y: 1 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1 }}
          className="text-4xl md:text-5xl font-extrabold text-gray-900 tracking-tight"
        >
          Welcome To <span className="text-green-600">Honor App</span>
        </motion.h1>

        {/* Subtitle Typing */}
        <motion.p
          animate={{ x: offset.x * 0.6, y: offset.y * 0.6 }}
          className="mt-3 text-lg text-gray-500 h-[28px]"
        >
          <Typewriter
            words={[
              "Sistem Perhitungan Honor Dokter RS Murni Teguh Susanna Wesley",
            ]}
            loop={0}
            cursor
            cursorStyle="_"
            typeSpeed={70}
            deleteSpeed={40}
            delaySpeed={2000}
          />
        </motion.p>

        {/* Premium Line Accent */}
        <motion.div
          initial={{ scaleX: 0 }}
          animate={{ scaleX: 1 }}
          transition={{ duration: 0.7, ease: "easeIn", delay: 0.4 }}
          className="mx-auto mt-5 w-96 h-[3px] bg-gradient-to-r from-green-600 to-emerald-400 rounded-full origin-left"
          style={{ animation: "breathing 4s ease-in-out infinite" }}
        />

        <style>
          {`
@keyframes breathing {
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(1.8); }
}
`}
        </style>
      </div>
    </motion.div>
  );
}

export default function Dashboard() {
  const [doctorname, setDoctor] = useState("");
  const [selectedDoctor, setSelectedDoctor] = useState(""); // nama dokter yang dipilih
  const [doctorList, setDoctorList] = useState([]);
  const [filteredDoctors, setFilteredDoctors] = useState([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const [totaldata, settotaldata] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const currentYear = new Date().getFullYear();

  const [lineData, setLineData] = useState([]);
  const years = Array.from({ length: 4 }, (_, i) => currentYear - i);
  const [year, setYear] = useState(String(currentYear));
  const [isOpen, setIsOpen] = useState(false);
  const [isOpena, setIsOpena] = useState(false);
  const [isOpenb, setIsOpenb] = useState(false);
  const months = [
    { label: "Januari", value: 1 },
    { label: "Februari", value: 2 },
    { label: "Maret", value: 3 },
    { label: "April", value: 4 },
    { label: "Mei", value: 5 },
    { label: "Juni", value: 6 },
    { label: "Juli", value: 7 },
    { label: "Agustus", value: 8 },
    { label: "September", value: 9 },
    { label: "Oktober", value: 10 },
    { label: "November", value: 11 },
    { label: "Desember", value: 12 },
  ];
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [selectedMonth, setSelectedMonth] = useState<number | null>(null);
  const [data, setData] = useState<HonorDokter[]>([]);
  const [filters, setFilters] = useState<{
    txn_doctor: string;
    month?: number | null;
    year?: number | null;
  }>({
    txn_doctor: "",
    month: null,
    year: currentYear,
  });

  useEffect(() => {
    const fetchDoctors = async () => {
      const res = await axios.get("http://localhost:8080/api/get-doctor-data", {
        withCredentials: true,
      });
      setDoctorList(
        res.data.data.map((d: any) => ({
          name: d.doctor_name || d.DoctorName || d.doctorName || d.Name || "",
          raw: d,
        }))
      );
    };
    fetchDoctors();
  }, []);

  // Fetch data dari API
  const fetchHonorChart = async () => {
    try {
      let url = `http://localhost:8080/api/honor-chart?year=${year}`;

      // fetch hanya jika dokter telah dipilih
      if (selectedDoctor !== "") {
        url += `&doctor_name=${selectedDoctor}`;
      }

      const res = await axios.get(url, { withCredentials: true });
      setLineData(res.data.monthlyHonor);
    } catch (err) {
      console.log("Error fetching chart:", err);
    }
  };

  // Auto reload setiap kali doctor/year berubah
  useEffect(() => {
    fetchHonorChart();
  }, [selectedDoctor, year]);

  // 🔄 Fetch data dari backend
  const fetchData = async (page = 1) => {
    try {
      const params: any = { txn_doctor: filters.txn_doctor, page };

      // Kirim bulan/tahun hanya jika user memilih
      if (filters.month) params.month = filters.month;
      if (filters.year) params.year = filters.year;

      const res = await axios.get(
        "http://localhost:8080/api/get-doctor-honor-monthly",
        {
          params,
          withCredentials: true,
        }
      );

      setData(res.data.data || []);
      setTotalPages(res.data.totalPages || 1);
      setCurrentPage(res.data.page || 1);
      settotaldata(res.data.total || 0);
    } catch (err) {
      console.error(err);
      Swal.fire("Error", "Gagal mengambil data honor dokter.", "error");
    }
  };

  useEffect(() => {
    fetchData(1); // fetch awal: semua data
  }, []);

  const handleSearch = () => {
    setCurrentPage(1);
    fetchData(1);
  };
  return (
    <div className="ml-64 mt-12 p-8 min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 focus:outline-none">
      <FuturisticHeader />
      <div className="xl:col-span-2 bg-white p-6 rounded-2xl shadow">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-lg font-semibold text-gray-700 mb-4">
            Tren Honor Per Bulan
          </h2>
          {selectedDoctor && (
            <span className="text-green-600">— {selectedDoctor} —</span>
          )}
          <div className="flex justify-between gap-2 ">
            {/* Search Dokter */}
            <div className=" relative  w-64">
              <input
                type="text"
                className="border rounded-xl p-2 w-full bg-transparent focus:outline-none focus:ring-2 focus:ring-green-400 text-gray-600"
                placeholder="Cari Dokter..."
                value={doctorname}
                onChange={(e) => {
                  const value = e.target.value;
                  setDoctor(value);

                  if (value === "") {
                    setFilteredDoctors([]);
                    setShowDropdown(false);
                    return;
                  }

                  const filtered = doctorList.filter((d: any) =>
                    d.name.toLowerCase().includes(value.toLowerCase())
                  );

                  setFilteredDoctors(filtered);
                  setShowDropdown(true);
                }}
                onFocus={() => {
                  if (doctorname !== "") setShowDropdown(true);
                  console.log("Dropdown dibuka");
                }}
                onBlur={() => setTimeout(() => setShowDropdown(false), 150)}
              />

              {/* Dropdown */}
              {showDropdown && (
                <div className="absolute top-full left-0 w-full bg-white text-gray-600 shadow-lg rounded-xl mt-1 max-h-48 overflow-y-auto z-10 border border-gray-200">
                  {filteredDoctors.length === 0 ? (
                    <div className="p-2 text-gray-400 text-sm">
                      Tidak ditemukan
                    </div>
                  ) : (
                    filteredDoctors.map((d: any, idx: number) => (
                      <div
                        key={idx}
                        className="p-2 cursor-pointer hover:bg-green-50 transition"
                        onClick={() => {
                          setDoctor(d.name);
                          setSelectedDoctor(d.name);
                          setShowDropdown(false);
                        }}
                      >
                        {d.name}
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>
            {/* Tombol Clear */}
            {doctorname !== "" && (
              <button
                onClick={() => {
                  setDoctor("");
                  setSelectedDoctor("");
                  setFilteredDoctors([]);
                  setShowDropdown(false);
                  fetchHonorChart();
                }}
                className="px-3 py-2 rounded-xl bg-red-500 text-white hover:bg-red-600 transition"
              >
                Clear
              </button>
            )}
            {/* Year Selector */}

            <div className="relative w-52">
              <button
                type="button"
                onClick={() => setIsOpenb(!isOpenb)}
                className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
      bg-white text-gray-700 shadow-sm transition-all duration-300 hover:border-green-400 
      focus:ring-2 focus:ring-green-300 focus:outline-none
      ${isOpenb ? "ring-2 ring-green-300" : ""}`}
              >
                <span className={year ? "text-gray-700" : "text-gray-400"}>
                  {year || "Year"}
                </span>

                <ChevronDown
                  size={18}
                  className={`transition-transform duration-300 ${
                    isOpena ? "rotate-180 text-green-500" : "text-gray-400"
                  }`}
                />
              </button>

              {isOpenb && (
                <div className="absolute text-gray-600 z-10 w-full mt-2 bg-white border border-gray-200 rounded-xl shadow-lg animate-fadeIn backdrop-blur-md">
                  {years.map((y) => (
                    <div
                      key={y}
                      onClick={() => {
                        setYear(String(y)); // update year utama
                        setIsOpena(false);
                      }}
                      className={`px-4 py-2 cursor-pointer transition-all duration-200 hover:bg-green-50 hover:text-green-600 ${
                        selectedYear == y
                          ? "bg-green-100 text-green-700 font-semibold"
                          : ""
                      }`}
                    >
                      {y}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>

        <ResponsiveContainer width="100%" height={300}>
          <LineChart
            data={lineData}
            margin={{ top: 20, right: 30, left: 60, bottom: 20 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="month" />
            <YAxis tickFormatter={(value) => value.toLocaleString("id-ID")} />
            <Tooltip
              formatter={(value) => value.toLocaleString("id-ID")}
              labelFormatter={(label) => `Bulan: ${label}`}
            />

            <Line
              type="monotone"
              dataKey="Total"
              stroke="#10B981"
              strokeWidth={3}
              dot={{ r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
      {/* Tabel Honor Per Bulan */}
      <div className="mt-8 relative z-[10] bg-white/80 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 hover:shadow-xl transition-all duration-300">
        <h2 className="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
          Data Honor Dokter Bulanan
        </h2>

        <div className="flex flex-wrap gap-4 items-center justify-between text-gray-600">
          <div className="flex flex-wrap gap-2 items-center">
            <input
              placeholder="Doctor Name..."
              value={filters.txn_doctor}
              onChange={(e) =>
                setFilters({ ...filters, txn_doctor: e.target.value })
              }
              className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
            />

            {/* Month Dropdown */}
            <div className="relative w-52">
              <button
                type="button"
                onClick={() => setIsOpen(!isOpen)}
                className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
             bg-white text-gray-700 shadow-sm transition-all duration-300 hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none
             ${isOpen ? "ring-2 ring-green-300" : ""}`}
              >
                <span className="text-gray-400">
                  {selectedMonth
                    ? months.find((m) => m.value === selectedMonth)?.label
                    : "Month"}
                </span>
                <ChevronDown
                  size={18}
                  className={`transition-transform duration-300 ${
                    isOpen ? "rotate-180 text-green-500" : "text-gray-400"
                  }`}
                />
              </button>

              {isOpen && (
                <div className="absolute z-10 w-full mt-2 bg-white border border-gray-200 rounded-xl shadow-lg animate-fadeIn backdrop-blur-md">
                  {months.map((month) => (
                    <div
                      key={month.value}
                      onClick={() => {
                        setSelectedMonth(month.value);
                        setFilters({
                          ...filters,
                          month: month.value,
                        });
                        setIsOpen(false);
                      }}
                      className={`px-4 py-2 cursor-pointer transition-all duration-200 hover:bg-green-50 hover:text-green-600 ${
                        selectedMonth === month.value
                          ? "bg-green-100 text-green-700 font-semibold"
                          : ""
                      }`}
                    >
                      {month.label}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Year Dropdown */}
            <div className="relative w-52">
              <button
                type="button"
                onClick={() => setIsOpena(!isOpena)}
                className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
             bg-white text-gray-700 shadow-sm transition-all duration-300 hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none
             ${isOpena ? "ring-2 ring-green-300" : ""}`}
              >
                <span className="text-gray-400">{selectedYear || "Year"}</span>
                <ChevronDown
                  size={18}
                  className={`transition-transform duration-300 ${
                    isOpena ? "rotate-180 text-green-500" : "text-gray-400"
                  }`}
                />
              </button>

              {isOpena && (
                <div className="absolute z-10 w-full mt-2 bg-white border border-gray-200 rounded-xl shadow-lg animate-fadeIn backdrop-blur-md">
                  {years.map((year) => (
                    <div
                      key={year}
                      onClick={() => {
                        setSelectedYear(year);
                        setFilters({ ...filters, year: year });
                        setIsOpena(false);
                      }}
                      className={`px-4 py-2 cursor-pointer transition-all duration-200 hover:bg-green-50 hover:text-green-600 ${
                        selectedYear === year
                          ? "bg-green-100 text-green-700 font-semibold"
                          : ""
                      }`}
                    >
                      {year}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-between gap-2">
            <button
              onClick={handleSearch}
              className="flex items-center gap-2 bg-gradient-to-r focus:outline-none from-green-500 to-green-600 text-white font-semibold px-5 py-2.5 rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
            >
              <Search className="w-4 h-4" /> Search
            </button>
          </div>
        </div>
        {/* Data Table */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <div className="overflow-x-auto rounded-xl border border-gray-200">
            <table className="w-full border-collapse text-sm">
              <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                <tr>
                  {["Doctor Name", "Honor"].map((head, i) => (
                    <th key={i} className="p-3 text-left font-semibold">
                      {head}
                    </th>
                  ))}
                </tr>
              </thead>

              <tbody>
                {data.length === 0 ? (
                  <tr>
                    <td
                      colSpan={9}
                      className="text-center py-6 text-gray-400 italic"
                    >
                      Tidak ada data ditemukan
                    </td>
                  </tr>
                ) : (
                  data.map((row, i) => (
                    <tr
                      key={i}
                      className="border-b border-gray-100 text-gray-600 hover:bg-green-50/50 transition-all duration-200"
                    >
                      <td className="p-3">{row.DoctorName}</td>
                      <td className="p-3 border text-right">
                        <div className="flex justify-end items-center gap-1">
                          <span>Rp</span>
                          <span>
                            {Number(Math.round(row.TotalHonor)).toLocaleString(
                              "id-ID"
                            )}
                          </span>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
        <div className="flex justify-between items-center gap-2 mt-6 text-gray-600 select-none ">
          {/* Tombol Prev */}
          <div className="text-sm text-gray-400">
            Jumlah Data: {totaldata.toLocaleString("id-ID")}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                if (currentPage > 1) fetchData(currentPage - 1);
              }}
              disabled={currentPage <= 1}
              className=" focus:outline-none focus:ring-0 outline-none hover:outline-none px-3 py-1.5 rounded-lg bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
            >
              <ChevronLeft />
            </button>

            {/* Nomor Halaman */}
            {Array.from({ length: totalPages }, (_, i) => i + 1)
              .filter(
                (page) =>
                  page === 1 ||
                  page === totalPages ||
                  (page >= currentPage - 1 && page <= currentPage + 1)
              )
              .map((page, i, arr) => (
                <React.Fragment key={page}>
                  {i > 0 && arr[i - 1] !== page - 1 && (
                    <span className="px-2">...</span>
                  )}
                  <button
                    onClick={() => fetchData(page)}
                    className={`px-3 py-1.5 rounded-lg transition-all ${
                      currentPage === page
                        ? "text-green-400 bg-transparent focus:outline-none focus:ring-0 outline-none"
                        : "text-gray-400 hover:text-gray-500 bg-transparent"
                    }`}
                  >
                    {page}
                  </button>
                </React.Fragment>
              ))}

            {/* Tombol Next */}
            <button
              onClick={() => {
                if (currentPage < totalPages) fetchData(currentPage + 1);
              }}
              disabled={currentPage >= totalPages}
              className="focus:outline-none focus:ring-0 outline-none px-3 py-1.5 rounded-lg bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
            >
              <ChevronRight />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

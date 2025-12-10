/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable react-hooks/rules-of-hooks */
import React, { useState, useEffect, useRef } from "react";
import axios from "axios";
import {
  Search,
  ChevronLeft,
  ChevronRight,
  X,
  ChevronDown,
} from "lucide-react";
import Swal from "sweetalert2";
import { AnimatePresence, motion } from "framer-motion";

// import Swal from "sweetalert2";

interface Honor {
  CardNo: number;
  RegnDept: string;
  WardDesc: string;
  TxnCategory: string;
  GlAccount: string;
  CareproviderTxnDoctorId: number;
  VisitNo: string;
  PatientName: string;
  PatientType: string;
  PatientClass: string;
  TxnCode: string;
  TxnDesc: string;
  TxnDoctor: string;
  RegnDoctor: string;
  RefDoctor: string;
  BasePrice: number;
  Qty: number;
  TxnAmount: number;
  MarginAmount: number;
  DiscountVisit: number;
  ClaimAmount: number;
  HonorLama: number;
  HonorBaru: number;
  Selisih: number;
  TarifINACBG: number;
  NetAmount: number;
  Status: string;
  BillDateTime: string;
  BillStatus: string;
  OrganisationName: string;
  AdmissionDateTime: string;
  DischargeDateTime: string;
}

export default function UpdateBilling() {
  const [data, setData] = useState<Honor[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [totaldata, settotaldata] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [doctorList, setDoctorList] = useState([]);
  const [doctorname, setDoctor] = useState("");
  const [selectedDoctor, setSelectedDoctor] = useState("");
  const [selectedDoctorID, setSelectedDoctorID] = useState("");
  const [selectedMonth, setSelectedMonth] = useState<number | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [filteredDoctors, setFilteredDoctors] = useState([]);
  const monthRef = useRef<HTMLDivElement>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [filters, setFilters] = useState({
    visit_no: "",
    txn_doctor: "",
  });
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    adjustment_value: "",
    counted_month: "",
    counted_year: "",
    note: "",
  });
  const [isOpen, setIsOpen] = useState(false);
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

  const handleChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    >
  ) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    // Validasi dasar
    if (
      !selectedDoctor ||
      !form.adjustment_value ||
      !selectedMonth ||
      !form.counted_year
    ) {
      Swal.fire(
        "Data belum lengkap",
        "Harap isi semua kolom wajib.",
        "warning"
      );
      return;
    }

    try {
      await axios.post(
        "http://localhost:8080/api/add-honor-adjustment",
        {
          doctor_name: selectedDoctor,
          adjustment_value: Number(form.adjustment_value),
          careprovider_txn_doctor_id: selectedDoctorID,
          counted_month: selectedMonth,
          counted_year: Number(form.counted_year),
          note: form.note,
        },
        { withCredentials: true }
      );

      Swal.fire("Berhasil!", "Penyesuaian honor berhasil disimpan.", "success");

      // Reset form
      setForm({
        adjustment_value: "",
        counted_month: "",
        counted_year: "",
        note: "",
      });
    } catch (err: any) {
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal menyimpan penyesuaian honor.",
        "error"
      );
    } finally {
      setShowForm(false);
    }
  };
  // 🔄 Fetch data dari backend
  const fetchData = async (page = 1) => {
    setLoading(true);
    setCurrentPage(page);

    try {
      const res = await axios.get(
        "http://localhost:8080/api/get-update-billing",
        {
          params: { ...filters, page },
          withCredentials: true,
        }
      );
      setData(res.data.data || []);
      setTotalPages(res.data.totalPages || 1);
      setCurrentPage(res.data.page || 1);
      settotaldata(res.data.total || 0);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData(1);
  }, []);
  // ketika klik search
  const handleSearch = () => {
    setCurrentPage(1);
    fetchData(1);
  };
  useEffect(() => {
    const fetchDoctors = async () => {
      const res = await axios.get("http://localhost:8080/api/get-doctor-list", {
        withCredentials: true,
      });
      setDoctorList(
        res.data.data.map((d: any) => ({
          name: d.doctor_name || d.DoctorName || d.doctorName || d.Name || "",
          careprovider_txn_doctor_id: d.CareproviderTxnDoctorId,
        }))
      );
    };
    fetchDoctors();
  }, []);
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (monthRef.current && !monthRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <>
      {showForm && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-xl shadow-lg w-[400px] relative">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold text-gray-700">
                Tambah Penyesuaian Honor
              </h2>
              <button
                className="bg-transparent text-red-400 focus:outline-none  hover:border-transparent hover:text-red-500"
                onClick={() => {
                  setShowForm(false);
                }}
              >
                <X>Back</X>
              </button>
            </div>
            {/* Nama Dokter */}
            <label className="text-gray-600 font-medium">Nama Dokter</label>
            <input
              type="text"
              className="border rounded-xl p-2 mb-4 w-full bg-transparent focus:outline-none focus:ring-2 focus:ring-green-400 text-gray-600"
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
            <AnimatePresence>
              {showDropdown && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  transition={{ duration: 0.15 }}
                >
                  <div className="absolute w-[88%] items-center bg-white text-gray-600 shadow-lg rounded-xl max-h-48 overflow-y-auto z-20 border border-gray-200 custom-scrollbar">
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
                            setSelectedDoctorID(d.careprovider_txn_doctor_id);
                          }}
                        >
                          {d.name}
                        </div>
                      ))
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Nominal Penyesuaian */}
            <label className="text-gray-600 font-medium">
              Nominal Penyesuaian
            </label>
            <input
              name="adjustment_value"
              type="number"
              value={form.adjustment_value}
              onChange={handleChange}
              placeholder="Masukkan nominal..."
              className="w-full p-2 text-gray-700 rounded-xl border focus:outline-none border-gray-300 mb-4 bg-white/60 focus:ring-2 focus:ring-green-400"
            />

            {/* Bulan */}
            <label className="text-gray-600 font-medium">Bulan</label>
            {/* Month Dropdown */}
            <div className="relative w-full" ref={monthRef}>
              <button
                type="button"
                onClick={() => {
                  setIsOpen(!isOpen);
                }}
                className={`w-full flex items-center justify-between mb-4 px-4 py-2 border border-gray-300 rounded-xl 
             bg-white text-gray-700 shadow-sm transition-all duration-300 hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none
             ${isOpen ? "ring-2 ring-green-300" : ""}`}
              >
                <span className="text-gray-400">
                  {selectedMonth
                    ? months.find((m) => m.value === selectedMonth)?.label
                    : "Pilih Bulan"}
                </span>
                <ChevronDown
                  size={18}
                  className={`transition-transform duration-300 ${
                    isOpen ? "rotate-180 text-green-500" : "text-gray-400"
                  }`}
                />
              </button>

              {isOpen && (
                <div className="absolute z-20 w-full bg-white border text-gray-400 border-gray-200 rounded-xl max-h-48 overflow-y-auto shadow-lg animate-fadeIn backdrop-blur-md custom-scrollbar">
                  {months.map((month) => (
                    <div
                      key={month.value}
                      onClick={() => {
                        setSelectedMonth(month.value);
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

            {/* Tahun */}
            <label className="text-gray-600 font-medium">Tahun</label>
            <input
              name="counted_year"
              type="number"
              value={form.counted_year}
              onChange={handleChange}
              placeholder="Masukkan tahun..."
              className="w-full p-2 text-gray-600 rounded-xl border focus:outline-none border-gray-300 mb-4 bg-white/60 focus:ring-2 focus:ring-green-400"
            />

            {/* Catatan */}
            <label className="text-gray-600 font-medium">
              Catatan (Opsional)
            </label>
            <textarea
              name="note"
              value={form.note}
              onChange={handleChange}
              placeholder="Tambahkan catatan..."
              className="w-full p-2 text-gray-600  rounded-xl border focus:outline-none border-gray-300 mb-4 bg-white/60 focus:ring-2 focus:ring-green-400 h-24 resize-none"
            />

            {/* Submit Button */}
            <button
              onClick={handleSubmit}
              className="w-full mt-2 py-2 focus:outline-none hover:border-transparent bg-green-600 hover:bg-green-700 text-white rounded-xl font-semibold shadow-md hover:shadow-lg transition-all duration-200"
            >
              Simpan Penyesuaian
            </button>
          </div>
        </div>
      )}

      {loading && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="relative w-20 h-20">
            {[...Array(8)].map((_, i) => {
              const angle = i * 60;
              return (
                <div
                  key={i}
                  className="absolute w-3 h-3 bg-green-400 rounded-full animate-spin-slow"
                  style={{
                    top: "50%",
                    left: "50%",
                    transform: `rotate(${angle}deg) translate(0, -30px)`,
                    animationDelay: `${i * 0.1}s`,
                  }}
                />
              );
            })}
          </div>

          <style>
            {`
          @keyframes spin-slow {
            0% { transform: rotate(0deg) translate(0, -30px); }
            100% { transform: rotate(360deg) translate(0, -30px); }
          }

          .animate-spin-slow {
            animation: spin-slow 1.5s linear infinite;
          }
        `}
          </style>
        </div>
      )}
      <div className="ml-64 mt-12 p-8 min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100">
        {/* =============== Section Upload =============== */}
        <div className="bg-white p-6 rounded-2xl shadow-sm border flex items-center justify-between">
          <div className="flex">
            <h2 className="text-xl font-semibold mb-2 text-gray-600">
              Proses Penyesuaian Honor
            </h2>
          </div>

          <div className="flex">
            <button
              onClick={() => setShowForm(true)} // ⬅️ buka form
              disabled={loading}
              className={`px-4 py-2 rounded-xl hover:border-transparent bg-blue-600 focus:outline-none text-white font-semibold shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200${
                loading ? "bg-gray-400" : " hover:bg-blue-700"
              }`}
            >
              {loading ? "Menghitung..." : "Penyesuaian Honor"}
            </button>
          </div>
        </div>

        {/* =============== Section Filter =============== */}
        <div className="mt-8 relative z-[10] bg-white/80 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 hover:shadow-xl transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
            Filter
          </h2>

          <div className="flex flex-wrap gap-4 items-center justify-between text-gray-600">
            {/* Input Filters */}
            <div className="flex flex-wrap gap-4 items-center">
              <input
                placeholder="Visit Number..."
                value={filters.visit_no}
                onChange={(e) =>
                  setFilters({ ...filters, visit_no: e.target.value })
                }
                className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
              />
              <input
                placeholder="Doctor Name..."
                value={filters.txn_doctor}
                onChange={(e) =>
                  setFilters({ ...filters, txn_doctor: e.target.value })
                }
                className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleSearch}
                className="flex items-center gap-2 hover:border-transparent focus:outline-none bg-gradient-to-r from-green-500 to-green-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
              >
                <Search className="w-4 h-4" />
                Search
              </button>
            </div>
          </div>
        </div>

        {/* =============== Section Data Table =============== */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700">
            Data Honor Update Billing
          </h2>

          {loading ? (
            <p className="text-gray-500 animate-pulse">Memuat data...</p>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-gray-200">
              <table className="w-full border-collapse text-sm">
                <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                  <tr>
                    {[
                      "Visit Number",
                      "Patient Name",
                      "Patient Class",
                      "Txn Code",
                      "Txn Desc",
                      "Txn Doctor",
                      "Honor Lama",
                      "Honor Baru",
                      "Selisih",
                    ].map((head, i) => (
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
                        className="border-b border-gray-100 text-gray-600  hover:bg-green-50/50 transition-all duration-200"
                      >
                        <td className="p-3">{row.VisitNo}</td>
                        <td className="p-3">{row.PatientName}</td>
                        <td className="p-3">{row.PatientClass}</td>
                        <td className="p-3">{row.TxnCode}</td>
                        <td className="p-3">{row.TxnDesc}</td>
                        <td className="p-3">{row.TxnDoctor}</td>
                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(Math.round(row.HonorBaru)).toLocaleString(
                                "id-ID"
                              )}
                            </span>
                          </div>
                        </td>

                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(Math.round(row.HonorLama)).toLocaleString(
                                "id-ID"
                              )}
                            </span>
                          </div>
                        </td>

                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(Math.round(row.Selisih)).toLocaleString(
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
          )}

          {/* Pagination */}
          <div className="flex justify-between items-center mt-6 text-gray-600 select-none">
            {/* Tombol Prev */}
            <div className="text-sm text-gray-400">
              Jumlah Data: {totaldata.toLocaleString("id-ID")}
            </div>
            <div className="flex items-center gap-2 focus:outline-none focus:ring-0 focus:ring-offset-0 ">
              <button
                onClick={() => {
                  if (currentPage > 1) fetchData(currentPage - 1);
                }}
                disabled={currentPage <= 1}
                className="px-3  rounded-lg bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
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
                      className={`px-3  rounded-lg transition-all ${
                        currentPage === page
                          ? "text-green-400 bg-transparent"
                          : "text-gray-400 hover:text-gray-500 bg-transparent "
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
                className="px-3  rounded-lg bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
              >
                <ChevronRight />
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

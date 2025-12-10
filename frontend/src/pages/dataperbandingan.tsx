/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { useState, useEffect, useRef } from "react";
import axios from "axios";
import {
  Upload,
  Search,
  ChevronDown,
  ChevronRight,
  ChevronLeft,
} from "lucide-react";
import Swal from "sweetalert2";

interface Comparison {
  VisitNumber: string;
  Status: string;
  VisitNumberTujuan: string;
  TarifINACBG: number;
  TarifBeforeTopup: number;
}
const CHUNK_SIZE = 2 * 1024 * 1024; // 2MB
function randomId(len = 12) {
  return Array.from(crypto.getRandomValues(new Uint8Array(len)))
    .map((b) => (b % 36).toString(36))
    .join("");
}
interface UploadSectionProps {
  title: string;

  file?: File | null;
  onFileSelect: (file: File) => void;
  inputRef?: React.RefObject<HTMLInputElement | null>;
  inputId?: string;
}
const UploadSection: React.FC<UploadSectionProps> = ({
  title,
  file,
  onFileSelect,
  inputRef,
  inputId,
}) => {
  return (
    <div className="bg-white p-6 rounded-2xl shadow-sm border flex items-center justify-between">
      <div className="w-[25%]">
        <h2 className="text-xl font-semibold mb-2 text-gray-600">{title}</h2>
      </div>
      <div className="w-[50%]">
        <div className="border border-gray-300 rounded-md p-2 bg-gray-50 text-sm text-gray-700 truncate">
          {file ? file.name : "Belum ada file dipilih"}
        </div>
      </div>
      <div className="w-[20%]">
        <label className="cursor-pointer flex items-center justify-center gap-2 bg-green-500 text-white py-2 hover:border-transparent bg-gradient-to-r from-green-500 to-green-600 font-semibold px-5 rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200">
          <Upload size={20} />
          <span>Pilih File</span>
          <input
            id={inputId}
            ref={inputRef}
            type="file"
            accept=".xlsx"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) onFileSelect(file);
            }}
          />
        </label>
      </div>
    </div>
  );
};

export default function Comparison() {
  const [data, setData] = useState<Comparison[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [totaldata, settotaldata] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const [filters, setFilters] = useState({ visit_number: "", status: "" });
  const [loading, setLoading] = useState(false);
  const fileInputBRef = useRef<HTMLInputElement | null>(null);

  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState("");
  const dropdownRef = useRef<HTMLDivElement>(null);
  const options = ["All", "FIX", "OFFER (PENDING)", "SATU SEP"];

  // 🔄 Fetch data dari backend
  const fetchData = async (page = 1) => {
    setLoading(true);
    setCurrentPage(page);
    try {
      const res = await axios.get(
        "http://localhost:8080/api/get-comparisondata",
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
  // 📤 Upload file ke backend
  async function uploadcomparisondataFile(file: File) {
    setLoading(true);
    try {
      const fileId = `${file.name}-${Date.now()}-${randomId(6)}`;
      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

      for (let i = 0; i < totalChunks; i++) {
        const start = i * CHUNK_SIZE;
        const end = Math.min(file.size, start + CHUNK_SIZE);
        const chunk = file.slice(start, end);

        const form = new FormData();
        form.append("fileId", fileId);
        form.append("chunkIndex", String(i));
        form.append("totalChunks", String(totalChunks));
        form.append("fileName", file.name);
        form.append("chunk", chunk);

        const resp = await axios.post(
          "http://localhost:8080/api/upload-patientbill-chunk",
          form,
          {
            withCredentials: true, // ✅ ini wajib kalau pakai cookie JWT
            headers: {
              "Content-Type": "multipart/form-data",
            },
          }
        );
        console.log("✅ Upload success:", resp.data);
      }

      // Finalisasi
      const resp = await axios.post(
        "http://localhost:8080/api/upload-comparisondata-complete",
        { fileId, fileName: file.name },
        {
          withCredentials: true, // ⬅️ penting agar cookie JWT otomatis dikirim
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      console.log("✅ Upload complete:", resp.data);

      Swal.fire({
        title: "Upload Berhasil",
        text: `File ${file.name} berhasil di upload`,
        icon: "success",
        showConfirmButton: false,
        timer: 1000,
        width: "360px",
        customClass: {
          popup: "rounded-2xl shadow-lg p-4",
          title: "text-lg font-semibold text-gray-600",
          htmlContainer: "text-sm text-gray-600",
          confirmButton:
            "bg-green-500 hover:bg-green-600 text-white text-sm rounded-lg px-4 py-1.5",
        },
      });

      console.log("Processing started:", resp.data);
    } catch (error: any) {
      setLoading(false);
      Swal.fire({
        title: "Upload Gagal",
        text: error.message || "Terjadi kesalahan saat upload.",
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
    } finally {
      setLoading(false);
      setTimeout(() => 500);
      if (fileInputBRef.current) {
        fileInputBRef.current.value = "";
      }
    }
  }
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);
  return (
    <>
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
        {/* Section Upload */}
        <UploadSection
          title="Upload Data Perbandingan"
          onFileSelect={(file) => uploadcomparisondataFile(file)}
          inputRef={fileInputBRef}
          inputId="fileInputB"
        />
        {/* =================== SECTION FILTER =================== */}
        <div className="mt-8 relative z-[20] bg-white/80 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 hover:shadow-xl transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
            🔍 Filter Data
          </h2>
          <div className="flex flex-wrap gap-4 items-center justify-between text-gray-600">
            <div className="flex flex-wrap gap-4 items-center">
              <input
                placeholder="Cari visit number..."
                value={filters.visit_number}
                onChange={(e) =>
                  setFilters({ ...filters, visit_number: e.target.value })
                }
                className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
              />
              <div className="relative w-52 " ref={dropdownRef}>
                {/* Selected Box */}
                <button
                  type="button"
                  onClick={() => setIsOpen(!isOpen)}
                  className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
        bg-white text-gray-700 shadow-sm transition-all duration-300 
        hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none
        ${isOpen ? "ring-2 ring-green-300" : ""}`}
                >
                  <span>
                    {selected ? (
                      selected
                    ) : (
                      <span className="text-gray-400">Pilih Status...</span>
                    )}
                  </span>
                  <ChevronDown
                    size={18}
                    className={`transition-transform duration-300 ${
                      isOpen ? "rotate-180 text-green-500" : "text-gray-400"
                    }`}
                  />
                </button>

                {/* Dropdown Menu */}
                {isOpen && (
                  <div
                    className="absolute z-10 w-full mt-2 bg-white border border-gray-200 rounded-xl shadow-lg 
          animate-fadeIn backdrop-blur-md"
                  >
                    {options.map((opt) => (
                      <div
                        key={opt}
                        onClick={() => {
                          setSelected(opt);
                          setFilters({
                            ...filters,
                            status: opt === "All" ? "" : opt,
                          });
                          setIsOpen(false);
                        }}
                        className={`px-4 py-2 cursor-pointer transition-all duration-200 
              hover:bg-green-50 hover:text-green-600 ${
                selected === opt
                  ? "bg-green-100 text-green-700 font-semibold"
                  : ""
              }`}
                      >
                        {opt}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
            <button
              onClick={handleSearch}
              className="flex items-center gap-2 bg-gradient-to-r focus:outline-none hover:border-transparent from-green-500 to-green-600 text-white font-semibold px-5 py-2.5 rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
            >
              <Search className="w-4 h-4" />
              Search
            </button>
          </div>
        </div>

        {/* =================== SECTION DATA =================== */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700">
            Data Perbandingan
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
                      "Status",
                      "Visit Number Tujuan",
                      "Tarif Ina Cbg",
                      "Tarif Sebelum Top Up",
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
                        <td className="p-2 border text-center">
                          {row.VisitNumber}
                        </td>
                        <td className="p-2 border text-center">{row.Status}</td>
                        <td className="p-2 border text-center">
                          {row.VisitNumberTujuan}
                        </td>
                        <td className="p-2 border text-right">
                          <div className="grid grid-cols-[20px_1fr] justify-end">
                            <span className="inline-block">Rp</span>{" "}
                            <span className="inline-block">
                              {Number(row.TarifINACBG).toLocaleString("id-ID")}
                            </span>
                          </div>
                        </td>
                        <td className="p-2 border text-right">
                          <div className="grid grid-cols-[20px_1fr] justify-end">
                            <span className="inline-block">Rp</span>{" "}
                            <span className="inline-block">
                              {Number(row.TarifBeforeTopup).toLocaleString(
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
              {/* Pagination */}
              <div className="flex justify-between items-center mt-6 text-gray-600 select-none">
                {/* Tombol Prev */}
                <div className="text-sm text-gray-400 ml-2">
                  Jumlah Data: {totaldata.toLocaleString("id-ID")}
                </div>
                <div className="flex items-center gap-2">
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
                    className="px-3  rounded-lg bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
                  >
                    <ChevronRight />
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
}

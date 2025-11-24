/* eslint-disable @typescript-eslint/no-explicit-any */

import { useState, useEffect } from "react";
import axios from "axios";
import { Search, ChevronDown, FileDown } from "lucide-react";
import Swal from "sweetalert2";
import * as XLSX from "xlsx-js-style";
import { saveAs } from "file-saver";

interface HonorDokter {
  DoctorName: string;
  CareproviderTxnDoctorId: number;
  TotalHonor: number;
}

export default function HonorDokter() {
  const [data, setData] = useState<HonorDokter[]>([]);

  const [filters, setFilters] = useState<{
    txn_doctor: string;
    counted_month?: number | null;
    counted_year?: number | null;
  }>({
    txn_doctor: "",
    counted_month: null,
    counted_year: null,
  });

  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [isOpena, setIsOpena] = useState(false);

  const currentYear = new Date().getFullYear();
  const years = [currentYear, currentYear + 1];

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

  // State selected
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [selectedMonth, setSelectedMonth] = useState<number | null>(null);

  // 🔄 Fetch data dari backend
  const fetchData = async (page = 1) => {
    setLoading(true);

    try {
      const params: any = { txn_doctor: filters.txn_doctor, page };

      // Kirim bulan/tahun hanya jika user memilih
      if (filters.counted_month) params.month = filters.counted_month;
      if (filters.counted_year) params.year = filters.counted_year;

      const res = await axios.get(
        "http://localhost:8080/api/get-doctor-honor",
        {
          params,
          withCredentials: true,
        }
      );

      setData(res.data.data || []);
    } catch (err) {
      console.error(err);
      Swal.fire("Error", "Gagal mengambil data honor dokter.", "error");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData(1); // fetch awal: semua data
  }, []);

  const handleSearch = () => {
    fetchData(1);
  };

  const exportToExcel = async () => {
    setLoading(true);
    try {
      const params: any = { txn_doctor: filters.txn_doctor, all: true };
      if (filters.counted_month) params.month = filters.counted_month;
      if (filters.counted_year) params.year = filters.counted_year;

      const res = await axios.get(
        "http://localhost:8080/api/get-doctor-honor",
        {
          params,
          withCredentials: true,
        }
      );
      const allData = res.data.data || [];

      if (allData.length === 0) {
        Swal.fire("Oops!", "Tidak ada data untuk diexport.", "warning");
        return;
      }

      const exportData = allData.map((item: HonorDokter) => ({
        "Doctor Name": item.DoctorName,
        "Careprovider Txn Doctor": item.CareproviderTxnDoctorId,
        Honor: item.TotalHonor,
      }));

      const worksheet = XLSX.utils.json_to_sheet(exportData);
      const headerKeys = Object.keys(exportData[0]);

      headerKeys.forEach((_, index) => {
        const cellAddress = XLSX.utils.encode_cell({ r: 0, c: index });
        const cell = worksheet[cellAddress];
        if (!cell) return;
        cell.s = {
          fill: { fgColor: { rgb: "C6EFCE" } },
          font: { bold: true, color: { rgb: "006100" } },
          alignment: { horizontal: "center", vertical: "center" },
          border: {
            top: { style: "thin", color: { rgb: "006100" } },
            bottom: { style: "thin", color: { rgb: "006100" } },
            left: { style: "thin", color: { rgb: "006100" } },
            right: { style: "thin", color: { rgb: "006100" } },
          },
        };
      });

      const columnWidths = headerKeys.map((key) => ({
        wch: Math.max(key.length + 2, 15),
      }));
      worksheet["!cols"] = columnWidths;

      const workbook = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(workbook, worksheet, "Honor Data");

      const excelBuffer = XLSX.write(workbook, {
        bookType: "xlsx",
        type: "array",
      });

      const file = new Blob([excelBuffer], {
        type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;charset=UTF-8",
      });
      saveAs(
        file,
        `Honor-Dokter_${new Date().toISOString().slice(0, 10)}.xlsx`
      );
    } catch (error) {
      setLoading(false);
      console.error(error);
      Swal.fire("Error", "Gagal export data.", "error");
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitRequest = async () => {
    if (data.length === 0) {
      Swal.fire("Oops!", "Tidak ada data untuk disubmit.", "warning");
      return;
    }
    setLoading(true);
    try {
      await axios.post(
        "http://localhost:8080/api/honor/submit-request",
        {
          counted_month: filters.counted_month,
          counted_year: filters.counted_year,
          data,
        },
        { withCredentials: true }
      );
      Swal.fire("Berhasil!", "Permohonan berhasil dikirim.", "success");
    } catch (err: any) {
      setLoading(false);
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal mengirim permohonan.",
        "error"
      );
    } finally {
      setLoading(false);
    }
  };

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
        {/* Filter Section */}

        <div className="mt-8 relative z-[10] bg-white/80 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 hover:shadow-xl transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
            🔍 Filter
          </h2>

          <div className="flex flex-wrap gap-4 items-center justify-between text-gray-600">
            <div className="flex flex-wrap gap-4 items-center">
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
                  <span>
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
                            counted_month: month.value,
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
                  <span>{selectedYear || "Year"}</span>
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
                          setFilters({ ...filters, counted_year: year });
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
                onClick={handleSubmitRequest}
                className="bg-gradient-to-r from-green-500 to-emerald-600 text-white px-5 py-2.5 rounded-xl shadow hover:scale-105 transition"
              >
                Submit Data Honor
              </button>

              <button
                onClick={exportToExcel}
                className="flex items-center gap-2 bg-gradient-to-r from-green-500 to-green-600 text-white font-semibold px-5 py-2.5 rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
              >
                <FileDown className="w-4 h-4" /> Export
              </button>

              <button
                onClick={handleSearch}
                className="flex items-center gap-2 bg-gradient-to-r from-green-500 to-green-600 text-white font-semibold px-5 py-2.5 rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
              >
                <Search className="w-4 h-4" /> Search
              </button>
            </div>
          </div>
        </div>

        {/* Data Table */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700">
            Data Honor
          </h2>

          {loading ? (
            <p className="text-gray-500 animate-pulse">Memuat data...</p>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-gray-200">
              <table className="w-full border-collapse text-sm">
                <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                  <tr>
                    {["Careprovider Txn Doctor ID", "Doctor Name", "Honor"].map(
                      (head, i) => (
                        <th key={i} className="p-3 text-left font-semibold">
                          {head}
                        </th>
                      )
                    )}
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
                        <td className="p-3">{row.CareproviderTxnDoctorId}</td>
                        <td className="p-3">{row.DoctorName}</td>
                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(
                                Math.round(row.TotalHonor)
                              ).toLocaleString("id-ID")}
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
        </div>
      </div>
    </>
  );
}

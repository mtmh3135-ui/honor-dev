/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable react-hooks/rules-of-hooks */
import React, { useState, useEffect, useRef } from "react";
import axios from "axios";
import {
  Search,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  FileDown,
} from "lucide-react";
import Swal from "sweetalert2";
import * as XLSX from "xlsx-js-style";
import { saveAs } from "file-saver";
// import Swal from "sweetalert2";

interface Honor {
  CardNo: number;
  RegnDept: string;
  WardDesc: string;
  TxnCategory: string;
  GlAccount: string;
  CareproviderTxnDoctorId: number;
  VisitNo: string;
  VisitNoFix: string;
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
  HonorMaster: number;
  HonorProp: number;
  HonorFinal: number;
  TarifINACBG: number;
  NetAmount: number;
  Status: string;
  BillDateTime: string;
  BillStatus: string;
  OrganisationName: string;
  AdmissionDateTime: string;
  DischargeDateTime: string;
}

export default function Honor() {
  const [data, setData] = useState<Honor[]>([]);
  const [totaldata, settotaldata] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const monthRef = useRef<HTMLDivElement>(null);
  const yearRef = useRef<HTMLDivElement>(null);
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
  const currentYear = new Date().getFullYear();
  const years = Array.from({ length: 4 }, (_, i) => currentYear - i);
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [selectedMonth, setSelectedMonth] = useState<number | null>(null);
  const [filters, setFilters] = useState<{
    visit_no: string;
    visit_no_fix: string;
    patient_name: string;
    patient_class: string;
    counted_month: number | null;
    counted_year: number | null;
  }>({
    visit_no: "",
    visit_no_fix: "",
    patient_name: "",
    patient_class: "",
    counted_month: null,
    counted_year: currentYear,
  });
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState("");
  const options = ["All", "BPJS", "General", "Insurance", "Corporate"];

  // 🔄 Fetch data dari backend
  const fetchData = async (page = 1) => {
    setLoading(true);
    setCurrentPage(page);

    try {
      const res = await axios.get("http://localhost:8080/api/get-honor-data", {
        params: { ...filters, page },
        withCredentials: true,
      });
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

  // Perhitungan Honor
  const honorcount = async () => {
    setLoading(true);

    try {
      const res = await axios.post(
        "http://localhost:8080/api/honor-count",
        {},
        {
          withCredentials: true,
        }
      );
      Swal.fire({
        icon: "success",
        title: "Selesai!",
        text: res.data.message,
        timer: 2000,
        showConfirmButton: false,
      });
    } catch (err: any) {
      Swal.fire({
        icon: "error",
        title: "Gagal",
        text: err.message || "Terjadi kesalahan",
      });
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

  const exportToExcel = async () => {
    try {
      const res = await axios.get("http://localhost:8080/api/get-honor-data", {
        params: { ...filters, all: true },
        withCredentials: true,
      });
      const allData = res.data.data || [];
      if (!data || allData.length === 0) {
        Swal.fire("Oops!", "Tidak ada data untuk diexport.", "warning");
        return;
      }
      const exportData = allData.map((item: Honor) => ({
        "Patient Name": item.PatientName,
        "Card No": item.CardNo,
        "Patient Type": item.PatientType,
        "Visit No": item.VisitNo,
        "Visit No Fix": item.VisitNoFix,
        "Regn Dept": item.RegnDept,
        "Ward Desc": item.WardDesc,
        "Patient Class": item.PatientClass,
        "Txn Category": item.TxnCategory,
        "Txn Code": item.TxnCode,
        "GL Account": item.GlAccount,
        "Txn Description": item.TxnDesc,
        "Careprovider Txn Doctor": item.CareproviderTxnDoctorId,
        "Txn Doctor": item.TxnDoctor,
        "Regn Doctor": item.RegnDoctor,
        "Ref Doctor": item.RefDoctor,
        "Base Price": item.BasePrice,
        Qty: item.Qty,
        "Txn Amount": item.TxnAmount,
        "Margin Amount": item.MarginAmount,
        "Claim Amount": item.ClaimAmount,
        "Discount Visit": item.DiscountVisit,
        "Honor Master": Math.round(item.HonorMaster),
        "Honor Prop": Math.round(item.HonorProp),
        "Honor Final": Math.round(item.HonorFinal),
        "Tarif Ina Cbg": Math.round(item.TarifINACBG),
        "Net Amount": Math.round(item.NetAmount),
        "Status Pembayaran BPJS": item.Status,
        "Bill DateTime": item.BillDateTime,
        "Bill Status": item.BillStatus,
        "Organisation Name": item.OrganisationName,
        "Admission Date Time": item.AdmissionDateTime,
        "Discharge Date Time": item.DischargeDateTime,
      }));

      //  Buat worksheet
      const worksheet = XLSX.utils.json_to_sheet(exportData);

      //  Ambil nama kolom (header)
      const headerKeys = Object.keys(data[0]);

      //  Styling header
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

      //  Otomatis set lebar kolom
      const columnWidths = headerKeys.map((key) => ({
        wch: Math.max(key.length + 2, 15), // minimal lebar 15
      }));
      worksheet["!cols"] = columnWidths;
      const workbook = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(workbook, worksheet, "Honor Data");

      //  Konversi ke file Excel
      const excelBuffer = XLSX.write(workbook, {
        bookType: "xlsx",
        type: "array",
      });

      //  Simpan file
      const file = new Blob([excelBuffer], {
        type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;charset=UTF-8",
      });
      saveAs(
        file,
        `Honor-Dokter_${new Date().toISOString().slice(0, 10)}.xlsx`
      );
    } catch {
      console.error(Error);
      Swal.fire("Error", "Gagal mengambil data untuk export.", "error");
    }
  };
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (monthRef.current && !monthRef.current.contains(e.target as Node)) {
        setIsOpenb(false);
      }
      if (yearRef.current && !yearRef.current.contains(e.target as Node)) {
        setIsOpena(false);
      }
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
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
        {/* =============== Section Upload =============== */}
        <div className="bg-white p-6 rounded-2xl shadow-sm border flex items-center justify-between">
          <div className="w-[25%]">
            <h2 className="text-xl font-semibold mb-2 text-gray-600">
              Proses Perhitungan Honor
            </h2>
          </div>

          <div className="w-[50%] text-gray-600">
            <span className="text-yellow-400 text-lg animate-pulse drop-shadow-[0_0_20px_rgba(250,204,21,1)]">
              Pastikan Data Patient Bill & Data Perbandingan Updated !
            </span>
          </div>
          <div className="w-[20%]">
            <button
              onClick={honorcount}
              disabled={loading}
              className={`px-4 py-2 rounded-xl hover:border-transparent focus:outline-none bg-blue-600 text-white font-semibold shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200${
                loading ? "bg-gray-400" : " hover:bg-blue-700"
              }`}
            >
              {loading ? "Menghitung..." : "Hitung Honor"}
            </button>
          </div>
        </div>

        {/* =============== Section Filter =============== */}
        <div className="mt-8 relative z-[10] bg-white/80 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 hover:shadow-xl transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
            🔍 Filter
          </h2>

          {/* Input Filters */}
          <div className="flex flex-wrap gap-5 items-center text-gray-500">
            <input
              placeholder="Visit Number..."
              value={filters.visit_no}
              onChange={(e) =>
                setFilters({ ...filters, visit_no: e.target.value })
              }
              className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
            />
            <input
              placeholder="Visit Number SEP..."
              value={filters.visit_no_fix}
              onChange={(e) =>
                setFilters({ ...filters, visit_no_fix: e.target.value })
              }
              className="border border-gray-300 focus:ring-2 focus:ring-green-400 p-2 rounded-xl w-60 bg-white/70 backdrop-blur-sm placeholder-gray-400 focus:outline-none"
            />

            <input
              placeholder="Patient Name..."
              value={filters.patient_name}
              onChange={(e) =>
                setFilters({ ...filters, patient_name: e.target.value })
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
                    <span className="text-gray-400">Patient Class</span>
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
                          patient_class: opt === "All" ? "" : opt,
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
            {/* Month Dropdown */}
            <div className="relative w-52" ref={monthRef}>
              <button
                type="button"
                onClick={() => setIsOpenb(!isOpenb)}
                className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
                           bg-white text-gray-700 shadow-sm transition-all duration-300 hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none
                           ${isOpenb ? "ring-2 ring-green-300" : ""}`}
              >
                <span className="text-gray-400">
                  {selectedMonth
                    ? months.find((m) => m.value === selectedMonth)?.label
                    : "Month"}
                </span>
                <ChevronDown
                  size={18}
                  className={`transition-transform duration-300 ${
                    isOpenb ? "rotate-180 text-green-500" : "text-gray-400"
                  }`}
                />
              </button>

              {isOpenb && (
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
                        setIsOpenb(false);
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
            <div className="relative w-52" ref={yearRef}>
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
          <div className="flex justify-end gap-2 mt-2">
            <button
              onClick={exportToExcel}
              className="flex items-center hover:border-transparent focus:outline-none gap-2 bg-gradient-to-r from-green-500 to-green-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
            >
              <FileDown className="w-4 h-4" />
              Export
            </button>
            <button
              onClick={handleSearch}
              className="flex items-center hover:border-transparent focus:outline-none gap-2 bg-gradient-to-r from-green-500 to-green-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200"
            >
              <Search className="w-4 h-4" />
              Search
            </button>
          </div>
        </div>

        {/* =============== Section Data Table =============== */}
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
                    {[
                      "Visit Number",
                      "Patient Name",
                      "Patient Class",
                      "Txn Code",
                      "Txn Desc",
                      "Txn Doctor",
                      "Honor Master",
                      "Honor Final",
                      "Tarif INA CBG",
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
                              {Number(
                                Math.round(row.HonorMaster)
                              ).toLocaleString("id-ID")}
                            </span>
                          </div>
                        </td>

                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(
                                Math.round(row.HonorFinal)
                              ).toLocaleString("id-ID")}
                            </span>
                          </div>
                        </td>

                        <td className="p-3 border text-right">
                          <div className="flex justify-end items-center gap-1">
                            <span>Rp</span>
                            <span>
                              {Number(
                                Math.round(row.TarifINACBG)
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

          {/* Pagination */}
          <div className="flex justify-between items-center mt-6 text-gray-600 select-none">
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
                className="px-3 rounded-lg bg-transparent text-gray-500 hover:text-gray-700 hover:border-transparent disabled:opacity-40 transition-all focus:outline-none focus:ring-0 outline-none hover:outline-none"
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
                          ? "text-green-400 bg-transparent focus:outline-none focus:ring-0 hover:border-transparent outline-none"
                          : "text-gray-400 hover:text-gray-500 bg-transparent hover:border-transparent "
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
                className="px-3 focus:outline-none focus:ring-0 rounded-lg hover:border-transparent bg-transparent text-gray-500 hover:text-gray-700 disabled:opacity-40 transition-all"
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

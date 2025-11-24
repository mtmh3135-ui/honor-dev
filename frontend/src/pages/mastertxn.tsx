/* eslint-disable @typescript-eslint/no-explicit-any */
import axios from "axios";
import React, { useEffect, useState } from "react";
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Search,
  Upload,
} from "lucide-react";
import Swal from "sweetalert2";
interface Mastertxn {
  TxnId: number;
  TxnCode: string;
  TxnDesc: string;
  TxnCategory: string;
  TxnType: string;
  BPJSIP: string;
  BPJSOP: string;
  RumusGeneral: string;
}
const UploadDoctor: React.FC = () => {
  const [file, setFile] = useState<File | null>(null);
  const fileInputRef = React.useRef<HTMLInputElement | null>(null);
  const [uploading, setUploading] = useState(false);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Mastertxn[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const [showAdd, setShowAdd] = useState(false);
  const [addData, setAddData] = useState({
    TxnCode: "",
    TxnDesc: "",
    TxnCategory: "",
    TxnType: "",
    BPJSIP: "",
    BPJSOP: "",
    RumusGeneral: "",
  });
  const resetForm = () => {
    setAddData({
      TxnCode: "",
      TxnDesc: "",
      TxnCategory: "",
      TxnType: "",
      BPJSIP: "",
      BPJSOP: "",
      RumusGeneral: "",
    });
  };
  const emptyTxn: Mastertxn = {
    TxnId: 0,
    TxnCode: "",
    TxnDesc: "",
    TxnCategory: "",
    TxnType: "",
    BPJSIP: "",
    BPJSOP: "",
    RumusGeneral: "",
  };
  const [editData, setEditData] = useState<Mastertxn>(emptyTxn);
  const [showEdit, setShowEdit] = useState(false);
  const [filters, setFilters] = useState({
    txn_code: "",
    txn_desc: "",
    txn_category: "",
  });
  const txnCategories = [
    "All",
    "ENT CLINIC",
    "SURGERY",
    "ADMINISTRATION",
    "ENDOSCOPY",
    "DENTAL CLINIC",
    "MATERNITY",
    "EQUIPMENT",
    "LIFESTYLE",
    "CONSULTATION",
    "EMERGENCY ROOM (ER)",
    "INTENSIVE CARE UNIT (ICU)",
    "HEMODIALYST",
    "OTHER",
    "PULMONOLOGY",
    "MEDICAL DIAGNOSIS",
    "UROLOGY",
  ];

  const txnTypes = ["visit", "fix", "tindakan"];
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState("");

  const fetchData = async (page = 1) => {
    setLoading(true);
    setCurrentPage(page);
    try {
      const res = await axios.get("http://localhost:8080/api/get-txn-data", {
        params: { ...filters, page },
        withCredentials: true,
      });
      setData(res.data.data || []);
      setTotalPages(res.data.totalPages || 1);
      setCurrentPage(res.data.page || 1);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    fetchData(1);
  }, []);

  const handleSearch = () => {
    setCurrentPage(1);
    fetchData(1);
  };

  const handleUpload = async () => {
    if (!file) {
      setMessage("Pilih file terlebih dahulu");
      return;
    }

    setUploading(true);
    setMessage("");

    try {
      const formData = new FormData();
      formData.append("file", file);

      const resp = await axios.post(
        "http://localhost:8080/api/upload-txn",
        formData,
        {
          withCredentials: true,
          headers: { "Content-Type": "multipart/form-data" },
        }
      );
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
      setFile(null);
      // RESET INPUT FILE
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      fetchData();
    } catch (error: any) {
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
      setUploading(false);
    }
  };
  const openEditModal = (row: Mastertxn) => {
    setEditData(row);
    setShowEdit(true);
  };
  const handleUpdate = async () => {
    try {
      await axios.put(
        `http://localhost:8080/api/update-txn/${editData.TxnId}`,
        {
          TxnCode: editData.TxnCode,
          TxnDesc: editData.TxnDesc,
          TxnCategory: editData.TxnCategory,
          TxnType: editData.TxnType,
          BPJSIP: editData.BPJSIP,
          BPJSOP: editData.BPJSOP,
          RumusGeneral: editData.RumusGeneral,
        },
        { withCredentials: true }
      );

      Swal.fire("Sukses", "Data berhasil diupdate", "success");

      setShowEdit(false);
      fetchData(currentPage);
    } catch (err: any) {
      Swal.fire({ text: err.message });
    }
  };
  const handleDelete = async (id: number) => {
    const confirm = await Swal.fire({
      title: "Hapus Data?",
      text: "Data tidak bisa dikembalikan!",
      icon: "warning",
      showCancelButton: true,
      confirmButtonText: "Hapus",
      cancelButtonText: "Batal",
    });

    if (!confirm.isConfirmed) return;

    try {
      await axios.delete(`http://localhost:8080/api/delete-txn/${id}`, {
        withCredentials: true,
      });

      Swal.fire("Terhapus", "Data sudah dihapus", "success");
      fetchData(currentPage);
    } catch (err: any) {
      Swal.fire({ text: err.message });
    }
  };

  const handleCreate = async () => {
    try {
      await axios.post(
        "http://localhost:8080/api/create-txn",
        {
          TxnCode: addData.TxnCode,
          TxnDesc: addData.TxnDesc,
          TxnCategory: addData.TxnCategory,
          TxnType: addData.TxnType,
          BPJSIP: addData.BPJSIP,
          BPJSOP: addData.BPJSOP,
          RumusGeneral: addData.RumusGeneral,
        },
        { withCredentials: true }
      );

      Swal.fire("Sukses", "Data berhasil ditambahkan", "success");
      resetForm();
      setShowAdd(false);
      fetchData(currentPage);
    } catch (err: any) {
      Swal.fire({ text: err.message });
    }
  };
  return (
    <>
      {showAdd && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-xl shadow-lg w-96 ">
            <h3 className="text-lg font-semibold mb-3 text-gray-600">
              Tambah Txn Baru
            </h3>
            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Txn Code"
              value={addData.TxnCode}
              onChange={(e) =>
                setAddData({ ...addData, TxnCode: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Txn Description"
              value={addData.TxnDesc}
              onChange={(e) =>
                setAddData({ ...addData, TxnDesc: e.target.value })
              }
            />

            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={addData.TxnCategory}
              onChange={(e) =>
                setAddData({ ...addData, TxnCategory: e.target.value })
              }
            >
              <option value="">Txn Category</option>
              {txnCategories.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={addData.TxnType}
              onChange={(e) =>
                setAddData({ ...addData, TxnType: e.target.value })
              }
            >
              <option value="">Txn Type</option>
              {txnTypes.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Honor BPJS Inpatients"
              value={addData.BPJSIP}
              onChange={(e) =>
                setAddData({ ...addData, BPJSIP: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Honor BPJS Outpatients"
              value={addData.BPJSOP}
              onChange={(e) =>
                setAddData({
                  ...addData,
                  BPJSOP: e.target.value,
                })
              }
            />
            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Honor Umum"
              value={addData.RumusGeneral}
              onChange={(e) =>
                setAddData({
                  ...addData,
                  RumusGeneral: e.target.value,
                })
              }
            />

            <div className="flex justify-end gap-2">
              <button
                onClick={() => {
                  setShowAdd(false);
                  resetForm();
                }}
                className="px-3 py-1 bg-gray-400 text-white rounded-lg"
              >
                Batal
              </button>

              <button
                onClick={handleCreate}
                className="px-3 py-1 bg-green-600 text-white rounded-lg"
              >
                Simpan
              </button>
            </div>
          </div>
        </div>
      )}
      {showEdit && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-xl shadow-lg w-96">
            <h3 className="text-lg font-semibold mb-3 text-gray-600">
              Edit Data Txn
            </h3>

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.TxnCode}
              onChange={(e) =>
                setEditData({ ...editData, TxnCode: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.TxnDesc}
              onChange={(e) =>
                setEditData({ ...editData, TxnDesc: e.target.value })
              }
            />
            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={editData.TxnCategory}
              onChange={(e) =>
                setEditData({ ...editData, TxnCategory: e.target.value })
              }
            >
              <option value="">-- Pilih Category --</option>
              {txnCategories.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={editData.TxnType}
              onChange={(e) =>
                setEditData({ ...editData, TxnType: e.target.value })
              }
            >
              <option value="">-- Pilih Category --</option>
              {txnTypes.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.BPJSIP}
              onChange={(e) =>
                setEditData({ ...editData, BPJSIP: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.BPJSOP}
              onChange={(e) =>
                setEditData({ ...editData, BPJSOP: e.target.value })
              }
            />
            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.RumusGeneral}
              onChange={(e) =>
                setEditData({ ...editData, RumusGeneral: e.target.value })
              }
            />

            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowEdit(false)}
                className="px-3 py-1 bg-gray-400 text-white rounded-lg"
              >
                Batal
              </button>

              <button
                onClick={handleUpdate}
                className="px-3 py-1 bg-blue-500 text-white rounded-lg"
              >
                Simpan
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="ml-64 mt-12 p-8 min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100">
        {/* === UPLOAD CARD === */}
        <div className="bg-white rounded-2xl shadow-md p-6 mb-8 border">
          <h2 className="text-xl font-bold mb-4 text-gray-700">
            Upload Data Dokter
          </h2>

          <div className="flex items-center gap-4">
            {/* FILE NAME BOX */}
            <div className="flex-1">
              <div className="border rounded-lg px-4 py-2 bg-gray-50 text-gray-600">
                {file ? file.name : "Belum ada file dipilih"}
              </div>

              {/* Progress Bar (ketika upload) */}
              {uploading && (
                <div className="w-full h-1 bg-gray-200 rounded mt-2 overflow-hidden">
                  <div className="h-full bg-green-500 animate-[progress_1s_infinite]"></div>
                </div>
              )}
            </div>

            {/* BUTTON UPLOAD */}
            <label className="bg-green-500 text-white px-4 py-2 rounded-lg cursor-pointer flex items-center gap-2 hover:bg-green-600">
              <Upload className="w-4 h-4" />
              Pilih File
              <input
                type="file"
                accept=".xlsx"
                ref={fileInputRef}
                onChange={(e) => setFile(e.target.files?.[0] || null)}
                className="hidden"
              />
            </label>

            <button
              onClick={handleUpload}
              disabled={!file || uploading}
              className={`px-4 py-2 rounded-lg text-white font-semibold ${
                uploading || !file
                  ? "bg-gray-400 cursor-not-allowed"
                  : "bg-green-600 hover:bg-green-700"
              }`}
            >
              {uploading ? "Uploading..." : "Upload"}
            </button>
          </div>

          {message && (
            <p className="mt-3 text-sm font-semibold text-green-600">
              {message}
            </p>
          )}
        </div>

        {/* === FILTER CARD === */}
        <div className="bg-white rounded-2xl shadow-md p-6 mb-8 border">
          <h2 className="text-lg font-bold mb-4 flex items-center gap-2 text-gray-600">
            🔍 Filter Data
          </h2>

          <div className="flex gap-4 items-center ">
            <input
              type="text"
              className="border rounded-lg px-4 py-2 flex-1 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Cari Deskripsi..."
              value={filters.txn_desc}
              onChange={(e) =>
                setFilters({ ...filters, txn_desc: e.target.value })
              }
            />
            <input
              type="text"
              className="border rounded-lg px-4 py-2 flex-1 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Cari Txn Code..."
              value={filters.txn_code}
              onChange={(e) =>
                setFilters({ ...filters, txn_code: e.target.value })
              }
            />
            <div className="relative w-52 ">
              {/* Selected Box */}
              <button
                type="button"
                onClick={() => setIsOpen(!isOpen)}
                className={`w-full flex items-center justify-between px-4 py-2 border border-gray-300 rounded-xl 
        bg-white text-gray-700 shadow-sm transition-all duration-300 
        hover:border-green-400 focus:ring-2 focus:ring-green-300 focus:outline-none  outline-none
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
                  className="absolute z-10 text-gray-400 w-full mt-2 bg-white border border-gray-200 rounded-xl shadow-lg 
          animate-fadeIn backdrop-blur-md"
                >
                  {txnCategories.map((opt) => (
                    <div
                      key={opt}
                      onClick={() => {
                        setSelected(opt);
                        setFilters({
                          ...filters,
                          txn_category: opt === "All" ? "" : opt,
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
            <button
              onClick={() => setShowAdd(true)}
              className="px-4 py-2 bg-green-600 text-white rounded-lg focus:outline-none"
            >
              Tambah Txn
            </button>
            <button
              onClick={handleSearch}
              className="flex items-center gap-2 bg-green-500 text-white px-4 py-2 rounded-lg hover:bg-green-700"
            >
              <Search className="w-4 h-4" /> Search
            </button>
          </div>
        </div>

        {/* === DATA TABLE === */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700">
            Data Dokter
          </h2>

          {loading ? (
            <p className="text-gray-500 animate-pulse">Memuat data...</p>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-gray-200">
              <table className="w-full border-collapse text-sm">
                <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                  <tr>
                    {[
                      "Txn Code",
                      "Txn Desc",
                      "Txn Category",
                      "Txn Type",
                      "Rumus BPJS IP",
                      "Rumus BPJS OP",
                      "Rumus Non BPJS",
                      "Aksi",
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
                        <td className="p-3">{row.TxnCode}</td>
                        <td className="p-3">{row.TxnDesc}</td>
                        <td className="p-3">{row.TxnCategory}</td>
                        <td className="p-3">{row.TxnType}</td>
                        <td className="p-3">{row.BPJSIP}</td>
                        <td className="p-3">{row.BPJSOP}</td>
                        <td className="p-3">{row.RumusGeneral}</td>
                        <td className="p-3 flex gap-2">
                          <button
                            onClick={() => openEditModal(row)}
                            className="px-3 py-1 bg-blue-500 text-white rounded-lg"
                          >
                            Edit
                          </button>

                          <button
                            onClick={() => handleDelete(row.TxnId)}
                            className="px-3 py-1 bg-red-500 text-white rounded-lg"
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          <div className="flex justify-between items-center gap-2 mt-6 text-gray-600 select-none focus:outline-none">
            {/* Tombol Prev */}
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

        <style>
          {`
          @keyframes progress {
            0% { width: 0%; }
            50% { width: 70%; }
            100% { width: 100%; }
          }
        `}
        </style>
      </div>
    </>
  );
};

export default UploadDoctor;

/* eslint-disable @typescript-eslint/no-explicit-any */
import axios from "axios";
import React, { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, Search, Upload } from "lucide-react";
import Swal from "sweetalert2";
interface Doctor {
  IdDoctor: number;
  DoctorName: string;
  Description: string;
  CareproviderTxnDoctorId: number;
}
const UploadDoctor: React.FC = () => {
  const [file, setFile] = useState<File | null>(null);
  const fileInputRef = React.useRef<HTMLInputElement | null>(null);
  const [uploading, setUploading] = useState(false);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Doctor[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const [showAdd, setShowAdd] = useState(false);
  const [addData, setAddData] = useState({
    DoctorName: "",
    Description: "",
    CareproviderTxnDoctorId: 0,
  });
  const resetForm = () => {
    setAddData({
      DoctorName: "",
      Description: "",
      CareproviderTxnDoctorId: 0,
    });
  };
  const emptyDoctor: Doctor = {
    IdDoctor: 0,
    DoctorName: "",
    Description: "",
    CareproviderTxnDoctorId: 0,
  };
  const [editData, setEditData] = useState<Doctor>(emptyDoctor);
  const [showEdit, setShowEdit] = useState(false);
  const [filters, setFilters] = useState({
    doctor_name: "",
    description: "",
  });

  const fetchData = async (page = 1) => {
    setLoading(true);
    setCurrentPage(page);
    try {
      const res = await axios.get("http://localhost:8080/api/get-doctor-data", {
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
        "http://localhost:8080/api/upload-doctor",
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
  const openEditModal = (row: Doctor) => {
    setEditData(row);
    setShowEdit(true);
  };
  const handleUpdate = async () => {
    try {
      await axios.put(
        `http://localhost:8080/api/update-doctor/${editData.IdDoctor}`,
        {
          DoctorName: editData.DoctorName,
          Description: editData.Description,
          CareproviderTxnDoctorId: editData.CareproviderTxnDoctorId,
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
      await axios.delete(`http://localhost:8080/api/delete-doctor/${id}`, {
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
        "http://localhost:8080/api/create-doctor",
        {
          DoctorName: addData.DoctorName,
          Description: addData.Description,
          CareproviderTxnDoctorId: addData.CareproviderTxnDoctorId,
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
              Tambah Dokter Baru
            </h3>
            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Nama Dokter"
              value={addData.DoctorName}
              onChange={(e) =>
                setAddData({ ...addData, DoctorName: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Deskripsi"
              value={addData.Description}
              onChange={(e) =>
                setAddData({ ...addData, Description: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="CareproviderTxnDoctorId"
              value={addData.CareproviderTxnDoctorId}
              onChange={(e) =>
                setAddData({
                  ...addData,
                  CareproviderTxnDoctorId: Number(e.target.value),
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
              Edit Data Dokter
            </h3>

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.DoctorName}
              onChange={(e) =>
                setEditData({ ...editData, DoctorName: e.target.value })
              }
            />

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.Description}
              onChange={(e) =>
                setEditData({ ...editData, Description: e.target.value })
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
        <div className="bg-white rounded-2xl shadow-md p-6 mb-8 border ">
          <h2 className="text-lg font-bold mb-4 flex items-center gap-2 text-gray-600">
            🔍 Filter Data
          </h2>

          <div className="flex gap-4 items-center ">
            <input
              type="text"
              className="border rounded-lg px-4 py-2 flex-1 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Cari Nama Dokter..."
              value={filters.doctor_name}
              onChange={(e) =>
                setFilters({ ...filters, doctor_name: e.target.value })
              }
            />
            <input
              type="text"
              className="border rounded-lg px-4 py-2 flex-1 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Cari Deskripsi..."
              value={filters.description}
              onChange={(e) =>
                setFilters({ ...filters, description: e.target.value })
              }
            />
            <button
              onClick={() => setShowAdd(true)}
              className="px-4 py-2 bg-green-600 text-white rounded-lg focus:outline-none"
            >
              Tambah Dokter
            </button>

            <button
              onClick={handleSearch}
              className="flex items-center gap-2 bg-green-500 text-white px-4 py-2 focus:outline-none rounded-lg hover:bg-green-700"
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
                      "Careprovider Txn Doctor Id",
                      "Doctor Name",
                      "Description",
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
                        <td className="p-3">{row.CareproviderTxnDoctorId}</td>
                        <td className="p-3">{row.DoctorName}</td>
                        <td className="p-3">{row.Description}</td>

                        <td className="p-3 flex gap-2">
                          <button
                            onClick={() => openEditModal(row)}
                            className="px-3 py-1 bg-blue-500 text-white rounded-lg"
                          >
                            Edit
                          </button>

                          <button
                            onClick={() => handleDelete(row.IdDoctor)}
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

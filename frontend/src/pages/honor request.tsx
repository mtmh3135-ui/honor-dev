/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from "react";
import axios from "axios";
import { CircleCheck, XCircle } from "lucide-react";
import Swal from "sweetalert2";
import { useNavigate } from "react-router-dom";

interface HonorRequestItem {
  id: number;
  description: string;
  counted_month: number;
  counted_year: number;
  status: string;
  created_at: string;
  username: string;
  approved_lvl1: string | null;
  approved_lvl2: string | null;
  cancelled_at: string | null;
}

export default function RequestList() {
  const [data, setData] = useState<HonorRequestItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [role, setRole] = useState<string>("");

  const navigate = useNavigate();

  const fetchData = async () => {
    try {
      const res = await axios.get(
        "http://localhost:8080/api/get-request-list",
        {
          withCredentials: true,
        }
      );
      setRole(res.data.role);
      setData(res.data.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleCancel = async (id: number) => {
    const confirm = await Swal.fire({
      title: "Batalkan Permohonan?",
      text: "Permohonan akan dibatalkan dan tidak dapat dipulihkan.",
      icon: "warning",
      showCancelButton: true,
      confirmButtonColor: "#d33",
      cancelButtonColor: "#3085d6",
      confirmButtonText: "Ya, batalkan!",
    });

    if (!confirm.isConfirmed) return;
    setLoading(true);
    await new Promise((r) => setTimeout(r, 50));
    try {
      await axios.put(
        `http://localhost:8080/api/honor-request/cancel/${id}`,
        {},
        { withCredentials: true }
      );

      Swal.fire("Dibatalkan", "Permohonan berhasil dibatalkan.", "success");
      fetchData();
      setTimeout(() => setLoading(false), 200);
    } catch (err: any) {
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal membatalkan permohonan.",
        "error"
      );
    }
  };

  const handleReject = async (id: number) => {
    const confirm = await Swal.fire({
      title: "Tolak Permohonan?",
      text: "Permohonan yang akan ditolak dan tidak dapat dipulihkan.",
      icon: "warning",
      showCancelButton: true,
      confirmButtonColor: "#d33",
      cancelButtonColor: "#3085d6",
      confirmButtonText: "Ya, Tolak!",
    });

    if (!confirm.isConfirmed) return;
    setLoading(true);
    await new Promise((r) => setTimeout(r, 50));
    try {
      await axios.put(
        `http://localhost:8080/api/honor-request/reject/${id}`,
        {},
        { withCredentials: true }
      );

      Swal.fire("Dibatalkan", "Permohonan berhasil dibatalkan.", "success");
      fetchData();
      setTimeout(() => setLoading(false), 200);
    } catch (err: any) {
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal membatalkan permohonan.",
        "error"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleApprovelvl1 = async (id: number) => {
    setLoading(true);
    try {
      await axios.put(
        `http://localhost:8080/api/honor/approve/1/${id}`,
        {},
        { withCredentials: true }
      );

      Swal.fire("Approved", "Permohonan berhasil di setujui.", "success");
      fetchData();
      setTimeout(() => setLoading(false), 200);
    } catch (err: any) {
      setLoading(false);
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal Setujui Permohonan.",
        "error"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleApprovelvl2 = async (id: number) => {
    setLoading(true);
    try {
      await axios.put(
        `http://localhost:8080/api/honor/approve/2/${id}`,
        {},
        { withCredentials: true }
      );

      Swal.fire("Approved", "Permohonan berhasil di setujui.", "success");
      fetchData();
      setTimeout(() => setLoading(false), 200);
    } catch (err: any) {
      setLoading(false);
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal Setujui Permohonan.",
        "error"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleApprovelvl3 = async (id: number) => {
    setLoading(true);
    try {
      await axios.put(
        `http://localhost:8080/api/honor/approve/3/${id}`,
        {},
        { withCredentials: true }
      );

      Swal.fire("Approved", "Permohonan berhasil di setujui.", "success");
      fetchData();
      setTimeout(() => setLoading(false), 200);
    } catch (err: any) {
      setLoading(false);
      Swal.fire(
        "Error",
        err.response?.data?.error || "Gagal Setujui Permohonan.",
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
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <h2 className="text-xl font-semibold mb-4 text-gray-700">
            Permohonan Honor Saya
          </h2>

          {loading ? (
            <p className="text-gray-500 animate-pulse ">Memuat data...</p>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-gray-200">
              <table className="w-full border-collapse text-sm">
                <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                  <tr>
                    {[
                      "Description",
                      "Bulan",
                      "Tahun",
                      "Status",
                      "Dibuat Oleh",
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
                        colSpan={10}
                        className="text-center py-6 text-gray-400 italic"
                      >
                        Tidak ada permohonan ditemukan
                      </td>
                    </tr>
                  ) : (
                    data.map((row, i) => (
                      <tr
                        key={i}
                        className="border-b border-gray-100 text-gray-600 hover:bg-green-50/50 transition-all duration-200 cursor-pointer"
                        onClick={() => {
                          navigate(`/request-list/${row.id}`);
                        }}
                      >
                        <td className="p-3">{row.description}</td>
                        <td className="p-3">{row.counted_month}</td>
                        <td className="p-3">{row.counted_year}</td>
                        <td className="p-3 font-semibold">
                          {row.status == "Pending_Approval_1" && (
                            <span className="text-blue-600">On Progress 1</span>
                          )}
                          {row.status == "Pending_Approval_2" && (
                            <span className="text-blue-600">On Progress 2</span>
                          )}
                          {row.status == "Pending_Approval_3" && (
                            <span className="text-blue-600">On Progress 3</span>
                          )}
                          {row.status == "Approved" && (
                            <span className="text-green-600">Approved</span>
                          )}
                          {row.status == "Rejected" && (
                            <span className="text-red-600">Rejected</span>
                          )}
                          {row.status == "Cancelled" && (
                            <span className="text-red-600">Cancelled</span>
                          )}
                        </td>
                        <td className="p-3">{row.username}</td>
                        <td
                          className="p-3"
                          onClick={(e) => e.stopPropagation()}
                        >
                          {/* USER → hanya bisa cancel */}
                          {role === "User" &&
                            row.status === "Pending_Approval_1" && (
                              <button
                                onClick={() => handleCancel(row.id)}
                                className="flex items-center gap-2 bg-red-500 text-white px-4 py-2 rounded-xl shadow hover:bg-red-600 transition-all duration-200 focus:outline-none hover:border-transparent"
                              >
                                <XCircle size={16} /> Cancel
                              </button>
                            )}
                          {/* HANDLING ADMIN */}
                          {role === "Admin" && (
                            <div className="flex justify-left items-center gap-3">
                              <button
                                onClick={() => handleReject(row.id)}
                                className="flex items-center gap-2 bg-red-500 text-white px-4 py-2 rounded-xl shadow hover:bg-red-600 transition-all duration-200 focus:outline-none hover:border-transparent"
                              >
                                <XCircle size={16} /> Reject
                              </button>
                            </div>
                          )}
                          {/* APPROVER LEVEL 1 */}
                          {role === "Approver_1" &&
                            row.status === "Pending_Approval_1" && (
                              <div className="flex justify-left items-center gap-3">
                                <button
                                  onClick={() => handleApprovelvl1(row.id)}
                                  className="flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-xl shadow hover:bg-green-700 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <CircleCheck size={16} />
                                  Approve
                                </button>
                                <button
                                  onClick={() => handleReject(row.id)}
                                  className="flex items-center gap-2 bg-red-500 text-white px-4 py-2 rounded-xl shadow hover:bg-red-600 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <XCircle size={16} /> Reject
                                </button>
                              </div>
                            )}

                          {/* APPROVER LEVEL 2 → hanya tampil jika sudah approved lvl1 */}
                          {role === "Approver_2" &&
                            row.status === "Pending_Approval_2" && (
                              <div className="flex justify-left items-center gap-3">
                                <button
                                  onClick={() => handleApprovelvl2(row.id)}
                                  className="flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-xl shadow hover:bg-green-700 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <CircleCheck size={16} />
                                  Approve
                                </button>
                                <button
                                  onClick={() => handleReject(row.id)}
                                  className="flex items-center gap-2 bg-red-500 text-white px-4 py-2 rounded-xl shadow hover:bg-red-600 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <XCircle size={16} /> Reject
                                </button>
                              </div>
                            )}
                          {/* APPROVER LEVEL 3 → hanya tampil jika sudah approved lvl2 */}
                          {role === "Approver_3" &&
                            row.status === "Pending_Approval_3" && (
                              <div className="flex justify-left items-center gap-3">
                                <button
                                  onClick={() => handleApprovelvl3(row.id)}
                                  className="flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-xl shadow hover:bg-green-700 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <CircleCheck size={16} />
                                  Approve
                                </button>
                                <button
                                  onClick={() => handleReject(row.id)}
                                  className="flex items-center gap-2 bg-red-500 text-white px-4 py-2 rounded-xl shadow hover:bg-red-600 transition-all duration-200 focus:outline-none hover:border-transparent"
                                >
                                  <XCircle size={16} /> Reject
                                </button>
                              </div>
                            )}
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

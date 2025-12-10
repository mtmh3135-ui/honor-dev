/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CircleCheck, X, XCircle } from "lucide-react";
import axios from "axios";
import Swal from "sweetalert2";

export default function HonorRequestDetail() {
  const { id } = useParams();
  const [detail, setDetail] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const [role, setRole] = useState<string>("");
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  useEffect(() => {
    fetchData();
  }, [id]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await axios.get(
        `http://localhost:8080/api/request-list/${id}`,
        { withCredentials: true }
      );
      setRole(res.data.role);
      setDetail(res.data);
    } catch (err) {
      console.error("Gagal memuat detail:", err);
      setDetail(null);
    } finally {
      setLoading(false);
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
      navigate(`/request-list`);
      fetchData();
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
      navigate(`/request-list`);
      fetchData();
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
      navigate(`/request-list`);
      fetchData();
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
      navigate(`/request-list`);
      fetchData();
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

  if (loading)
    return (
      <>
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
      </>
    );
  if (!detail) return <div>Data tidak ditemukan</div>;

  return (
    <div className="ml-64 mt-12 p-8 min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100">
      <div className="bg-white p-6 rounded-xl shadow-md flex justify-between">
        <h2 className="text-2xl font-bold mb-2 text-gray-600">
          Detail Permohonan Honor
        </h2>

        <div className="flex justify-between gap-2">
          {/* HANDLING ADMIN */}
          {role === "Admin" && (
            <div className="flex items-center gap-3">
              <button
                onClick={() => handleReject(detail.request.id)}
                className="flex items-center gap-2 bg-transparent border-red-500 text-gray-600 px-4 py-2 rounded-xl shadow hover:bg-red-500 transition-all duration-200 focus:outline-none"
              >
                <XCircle size={16} /> Reject
              </button>
            </div>
          )}
          {/* APPROVER LEVEL 1 */}
          {role === "Approver_1" &&
            detail.request.status === "Pending_Approval_1" && (
              <div className="flex items-center gap-3">
                <button
                  onClick={() => handleApprovelvl1(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-green-600 text-green-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-green-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <CircleCheck size={16} />
                  Approve
                </button>
                <button
                  onClick={() => handleReject(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-red-500 text-red-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-red-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <XCircle size={16} /> Reject
                </button>
              </div>
            )}

          {/* APPROVER LEVEL 2 → hanya tampil jika sudah approved lvl1 */}
          {role === "Approver_2" &&
            detail.request.status === "Pending_Approval_2" && (
              <div className="flex items-center gap-3">
                <button
                  onClick={() => handleApprovelvl2(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-green-600 text-green-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-green-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <CircleCheck size={16} />
                  Approve
                </button>
                <button
                  onClick={() => handleReject(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-red-500 text-red-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-red-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <XCircle size={16} /> Reject
                </button>
              </div>
            )}
          {/* APPROVER LEVEL 3 → hanya tampil jika sudah approved lvl2 */}
          {role === "Approver_3" &&
            detail.request.status === "Pending_Approval_3" && (
              <div className="flex items-center gap-3">
                <button
                  onClick={() => handleApprovelvl3(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-green-600 text-green-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-green-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <CircleCheck size={16} />
                  Approve
                </button>
                <button
                  onClick={() => handleReject(detail.request.id)}
                  className="flex items-center gap-2 bg-transparent border-red-500 text-red-600 hover:text-white px-4 py-2 rounded-xl shadow hover:bg-red-500 transition-all duration-200 focus:outline-none hover:border-transparent"
                >
                  <XCircle size={16} /> Reject
                </button>
              </div>
            )}
          <button
            className="bg-transparent text-red-400 focus:outline-none hover:border-transparent hover:text-red-500"
            onClick={() => {
              navigate(`/request-list`);
            }}
          >
            <X>Back</X>
          </button>
        </div>
      </div>

      {/* ===== TABLE DOCTOR SUMMARY ===== */}
      <div className="mt-6 bg-white p-6 rounded-xl shadow">
        <table className="w-full text-left border-collapse rounded-xl">
          <thead className="bg-green-500  text-white">
            <tr>
              <th className="p-3 font-semibold ">Nama Dokter</th>
              <th className="p-3 font-semibold ">Total Honor</th>
              <th className="p-3 font-semibold  text-center">Detail</th>
            </tr>
          </thead>

          <tbody>
            {detail.doctors?.map((doc: any, index: number) => (
              <React.Fragment key={index}>
                <tr className="border-b border-gray-100 text-gray-600 hover:bg-green-50">
                  <td className="p-3">{doc.doctor_name}</td>
                  <td className="p-3 font-bold text-green-700">
                    Rp {Math.round(doc.total_honor).toLocaleString("id-ID")}
                  </td>
                  <td className="p-3 text-center">
                    <button
                      onClick={() =>
                        setOpenIndex(openIndex === index ? null : index)
                      }
                      className="px-3 py-1 bg-green-500 text-white rounded-md text-sm focus:outline-none hover:border-transparent"
                    >
                      {openIndex === index ? "Tutup" : "Lihat"}
                    </button>
                  </td>
                </tr>

                {/* ===== RINCIAN DROPDOWN TABLE ===== */}
                {openIndex === index && (
                  <tr>
                    <td colSpan={3} className="bg-gray-50 p-4">
                      <table className="w-full text-sm border border-gray-300 rounded-lg overflow-hidden">
                        <thead>
                          <tr className="bg-green-500 text-white">
                            <th className="p-2 border"> </th>
                            <th className="p-2 border">Masuk</th>
                            <th className="p-2 border">Keluar</th>
                            <th className="p-2 border">Nomor RM</th>
                            <th className="p-2 border">Visit Number</th>
                            <th className="p-2 border">Nama Pasien</th>
                            <th className="p-2 border">Company</th>
                            <th className="p-2 border">Deskripsi</th>
                            <th className="p-2 border">Honor</th>
                          </tr>
                        </thead>
                        <tbody>
                          {doc.details?.map((v: any, vi: number) =>
                            v.items?.map((item: any, ii: number) => (
                              <tr
                                key={`${vi}-${ii}`}
                                className="border text-gray-600"
                              >
                                <td className="p-2 border">
                                  {item.patient_type}
                                </td>
                                <td className="p-2 border">{item.masuk}</td>
                                <td className="p-2 border">{item.keluar}</td>
                                <td className="p-2 border">{item.nrm}</td>
                                <td className="p-2 border">{v.visit_no}</td>
                                <td className="p-2 border">{item.pasien}</td>
                                <td className="p-2 border">{item.company}</td>
                                <td className="p-2 border">{item.txn_desc}</td>
                                <td className="p-2 border text-right">
                                  Rp {item.honor.toLocaleString("id-ID")}
                                </td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

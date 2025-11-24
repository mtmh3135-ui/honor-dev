/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { X } from "lucide-react";
import axios from "axios";

export default function HonorRequestDetail() {
  const { id } = useParams();
  const [detail, setDetail] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  useEffect(() => {
    fetchDetail();
  }, [id]);

  const fetchDetail = async () => {
    setLoading(true);
    try {
      const res = await axios.get(
        `http://localhost:8080/api/request-list/${id}`,
        { withCredentials: true }
      );

      setDetail(res.data);
    } catch (err) {
      console.error("Gagal memuat detail:", err);
      setDetail(null);
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
        <button
          className="bg-transparent text-red-600 focus:outline-none"
          onClick={() => {
            navigate(`/request-list`);
          }}
        >
          <X></X>
        </button>
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
                      className="px-3 py-1 bg-green-500 text-white rounded-md text-sm focus:outline-none"
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
                            <th className="p-2 border">Visit No</th>
                            <th className="p-2 border">Txn Code</th>
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
                                <td className="p-2 border">{v.visit_no}</td>
                                <td className="p-2 border">{item.txn_code}</td>
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

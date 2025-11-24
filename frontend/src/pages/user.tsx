/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";
import axios from "axios";

import Swal from "sweetalert2";
import { UserRoundPlus } from "lucide-react";

interface User {
  UserID: number;
  Username: string;
  Role: string;
  Password: string;
}

export default function UserManagement() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [showAdd, setShowAdd] = useState(false);
  const [addData, setAddData] = useState({
    Username: "",
    Role: "",
    Password: "",
  });
  const resetForm = () => {
    setAddData({
      Username: "",
      Role: "",
      Password: "",
    });
  };
  const emptyUser: User = {
    UserID: 0,
    Username: "",
    Role: "",
    Password: "",
  };
  const Roles = ["Admin", "User", "Approver_1", "Approver_2"];
  const [editData, setEditData] = useState<User>(emptyUser);
  const [showEdit, setShowEdit] = useState(false);
  // ✅ Ambil data user
  const fetchUsers = async () => {
    try {
      setLoading(true);
      const res = await axios.get("http://localhost:8080/api/users", {
        withCredentials: true,
      });
      setUsers(res.data.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const openEditModal = (row: User) => {
    setEditData(row);
    setShowEdit(true);
  };
  // ✅ Simpan user (Tambah / Edit)
  const handleUpdate = async () => {
    try {
      await axios.put(
        `http://localhost:8080/api/edit-user/${editData.UserID}`,
        {
          Username: editData.Username,
          Role: editData.Role,
          Password: editData.Password,
        },
        { withCredentials: true }
      );

      Swal.fire("Sukses", "User berhasil diupdate", "success");
      fetchUsers();
      setShowEdit(false);
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
      await axios.delete(`http://localhost:8080/api/delete-user/${id}`, {
        withCredentials: true,
      });

      Swal.fire("Terhapus", "User sudah dihapus", "success");
      fetchUsers();
    } catch (err: any) {
      Swal.fire({ text: err.message });
    }
  };

  const handleCreate = async () => {
    try {
      await axios.post(
        "http://localhost:8080/api/create-user",
        {
          Username: addData.Username,
          Password: addData.Password,
          Role: addData.Role,
        },
        { withCredentials: true }
      );

      Swal.fire("Sukses", "User berhasil ditambahkan", "success");
      resetForm();
      setShowAdd(false);
      fetchUsers();
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
              Tambah User Baru
            </h3>
            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Username"
              value={addData.Username}
              onChange={(e) =>
                setAddData({ ...addData, Username: e.target.value })
              }
            />

            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={addData.Role}
              onChange={(e) => setAddData({ ...addData, Role: e.target.value })}
            >
              <option value="">Role</option>
              {Roles.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <input
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent focus:outline-none"
              placeholder="Password"
              value={addData.Password}
              onChange={(e) =>
                setAddData({ ...addData, Password: e.target.value })
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
              Edit Data User
            </h3>

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.Username}
              onChange={(e) =>
                setEditData({ ...editData, Username: e.target.value })
              }
            />

            <select
              className="border px-3 py-2 w-full mb-3 text-gray-600 bg-transparent"
              value={editData.Role}
              onChange={(e) =>
                setEditData({ ...editData, Role: e.target.value })
              }
            >
              <option value="">-- Pilih Category --</option>
              {Roles.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>

            <input
              className="border px-3 py-2 w-full mb-3 bg-transparent text-gray-500"
              value={editData.Password}
              onChange={(e) =>
                setEditData({ ...editData, Password: e.target.value })
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
        {/* === DATA TABLE === */}
        <div className="mt-8 bg-white/90 backdrop-blur-lg p-6 rounded-2xl shadow-lg border border-gray-200 transition-all duration-300">
          <div className="flex justify-between">
            <h2 className="text-xl font-semibold mb-4 text-gray-700">
              Data User
            </h2>
          </div>

          {loading ? (
            <p className="text-gray-500 animate-pulse">Memuat data...</p>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-gray-200">
              <table className="w-full border-collapse text-sm">
                <thead className="bg-gradient-to-r from-green-500 to-green-600 text-white sticky top-0">
                  <tr>
                    <th className="p-3 text-left font-semibold w-[30%]">
                      Username
                    </th>
                    <th className="p-3 text-left font-semibold w-[15%]">
                      Role
                    </th>
                    <th className="p-3 text-left font-semibold w-[30%]">
                      Aksi
                    </th>
                    <th className=" p-3 text-left font-semibold w-[25%]">
                      <button
                        onClick={() => setShowAdd(true)}
                        className="flex justify-between gap-3 px-4 py-2 bg-green-100 text-green-700 hover:bg-green-200 rounded-lg focus:outline-none"
                      >
                        <UserRoundPlus className="w-4 h-4" />
                        Tambah User
                      </button>
                    </th>
                  </tr>
                </thead>

                <tbody>
                  {users.length === 0 ? (
                    <tr>
                      <td
                        colSpan={9}
                        className="text-center py-6 text-gray-400 italic"
                      >
                        Tidak ada data ditemukan
                      </td>
                    </tr>
                  ) : (
                    users.map((row, i) => (
                      <tr
                        key={i}
                        className="border-b border-gray-100 text-gray-600  hover:bg-green-50/50 transition-all duration-200"
                      >
                        <td className="p-3">{row.Username}</td>
                        <td className="p-3">{row.Role}</td>
                        <td className="p-3 flex gap-2">
                          <button
                            onClick={() => openEditModal(row)}
                            className="px-3 py-1 bg-blue-500 text-white rounded-lg"
                          >
                            Edit
                          </button>

                          <button
                            onClick={() => handleDelete(row.UserID)}
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
}

import Login from "./pages/login";
import Dashboard from "./pages/dashboard";

import { Navigate, Route, Routes } from "react-router-dom";
import ProtectedRoute from "./components/protectedRoute";
import PatientBill from "./pages/patientbill";
import DataPerbandingan from "./pages/dataperbandingan";
import Honor from "./pages/honor";
import Admin from "./pages/user";
import HonorDokter from "./pages/honordokter";
import Layout from "./layout/layout";
import Masterdoctor from "./pages/masterdoctor";
import Mastertxn from "./pages/mastertxn";
import Swal from "sweetalert2";
import AisData from "./pages/piutang";
import RequestList from "./pages/honor request";
import HonorRequestDetail from "./pages/honor detail";
import UpdateBilling from "./pages/updatebilling";

// Atur konfigurasi default agar tidak ubah layout
Swal.mixin({
  scrollbarPadding: false,
  heightAuto: false,
});

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/login" replace />} />
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="patient-bill" element={<PatientBill />} />
        <Route path="comparison" element={<DataPerbandingan />} />
        <Route path="adjustment" element={<UpdateBilling />} />
        <Route path="master-txn" element={<Mastertxn />} />
        <Route path="master-doctor" element={<Masterdoctor />} />
        <Route path="request-list" element={<RequestList />} />
        <Route path="request-list/:id" element={<HonorRequestDetail />} />
        <Route path="honor-data" element={<Honor />} />
        <Route path="honor-dokter" element={<HonorDokter />} />
        <Route path="data-piutang" element={<AisData />} />
        <Route path="admin" element={<Admin />} />
      </Route>
    </Routes>
  );
}

import Topbar from "../components/topbar";
import ModernSidebar from "../components/sidebar";

import { Outlet } from "react-router-dom";

export default function Layout() {
  return (
    <div className="flex w-screen bg-gray-50 ">
      {/* Sidebar */}
      <ModernSidebar />
      {/* Main Content */}
      <div className="flex-1 flex flex-col ">
        {/* Topbar */}
        <Topbar />
        {/* Content */}
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

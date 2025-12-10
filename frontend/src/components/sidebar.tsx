import React, { useState, useEffect } from "react";
import {
  LayoutDashboard,
  FileText,
  BadgeDollarSign,
  ShieldUser,
  Cog,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";
import axios from "axios";

type MenuItem = {
  label: string;
  icon: React.ReactNode;
  path?: string;
  children?: ChildItem[];
  allowedRoles: string[];
};
type ChildItem = {
  label: string;
  path?: string;
  children?: { label: string; path: string }[];
  allowedRoles: string[];
};

const menuItems: MenuItem[] = [
  {
    label: "Dashboard",
    icon: <LayoutDashboard className="w-5 h-5" />,
    path: "/dashboard",
    allowedRoles: ["Admin", "User", "Approver_1", "Approver_2", "Approver_3"],
  },
  {
    label: "Data",
    icon: <FileText className="w-5 h-5" />,
    allowedRoles: ["Admin", "User"],
    children: [
      {
        label: "Patient Bill",
        path: "/patient-bill",
        allowedRoles: ["Admin", "User"],
      },
      {
        label: "Data Piutang",
        path: "/data-piutang",
        allowedRoles: ["Admin", "User"],
      },
      {
        label: "Data Perbandingan",
        path: "/comparison",
        allowedRoles: ["Admin", "User"],
      },
      {
        label: "Penyesuaian Honor",
        path: "/adjustment",
        allowedRoles: ["Admin", "User"],
      },
    ],
  },

  {
    label: "Honor",
    icon: <BadgeDollarSign className="w-5 h-5" />,
    allowedRoles: ["Admin", "User", "Approver_1", "Approver_2", "Approver_3"],
    children: [
      {
        label: "Data Permohonan",
        path: "/request-list",
        allowedRoles: [
          "Admin",
          "User",
          "Approver_1",
          "Approver_2",
          "Approver_3",
        ],
      },
      {
        label: "Data Honor",
        path: "/honor-data",
        allowedRoles: [
          "Admin",
          "User",
          "Approver_1",
          "Approver_2",
          "Approver_3",
        ],
      },
      {
        label: "Honor Dokter",
        path: "/honor-dokter",
        allowedRoles: [
          "Admin",
          "User",
          "Approver_1",
          "Approver_2",
          "Approver_3",
        ],
      },
    ],
  },
  {
    label: "Master",
    icon: <Cog className="w-5 h-5" />,
    allowedRoles: ["Admin", "User"],
    children: [
      {
        label: "Master TXN",
        path: "/master-txn",
        allowedRoles: ["Admin", "User"],
      },
      {
        label: "Master Doctor",
        path: "/master-doctor",
        allowedRoles: ["Admin", "User"],
      },
    ],
  },
  {
    label: "User Management",
    icon: <ShieldUser className="w-5 h-5" />,
    path: "/admin",
    allowedRoles: ["Admin"],
  },
];

export default function SidebarCollapsible() {
  const navigate = useNavigate();
  const location = useLocation();
  const [activePath, setActivePath] = useState("");
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);
  const [role, setRole] = useState<string>("");
  useEffect(() => {
    axios
      .get("http://localhost:8080/api/me", { withCredentials: true })
      .then((res) => setRole(res.data.role))
      .catch(() => setRole(""));
  }, []);

  useEffect(() => {
    setActivePath(location.pathname);

    // buka dropdown jika anak aktif
    const activeParent = menuItems.find((item) =>
      item.children?.some((child) => child.path === location.pathname)
    );
    if (activeParent) setOpenDropdown(activeParent.label);
  }, [location.pathname]);

  const handleNavigate = (path?: string) => {
    if (path) navigate(path);
  };

  const toggleDropdown = (label: string) => {
    setOpenDropdown(openDropdown === label ? null : label);
  };
  // Filter menu sesuai role
  const filteredMenu = menuItems.filter(
    (menu) =>
      !menu.allowedRoles ||
      menu.allowedRoles.map((r) => r.toLowerCase()).includes(role.toLowerCase())
  );

  return (
    <aside
      className={`fixed top-0 left-0 h-full bg-white text-gray-600 shadow-xl flex flex-col w-64 `}
    >
      {/* Header */}
      <div className="px-6 py-6 border-none flex justify-center">
        <h1 className="text-2xl font-bold bg-clip-text text-gray-500">
          Honor App
        </h1>
      </div>

      {/* Menu */}
      <nav className="flex-1 px-3 py-4 overflow-y-auto space-y-2">
        {filteredMenu.map((item, index) => {
          const isDropdown = !!item.children;
          const isOpen = openDropdown === item.label;
          const isActive = activePath === item.path;

          return (
            <div key={index}>
              <button
                onClick={() =>
                  isDropdown
                    ? toggleDropdown(item.label)
                    : handleNavigate(item.path)
                }
                className={`group relative flex items-center justify-between w-full px-4 py-3 rounded-xl text-sm font-medium transition-all bg-transparent hover:border-transparent  focus:outline-none ${
                  isActive
                    ? " bg-gradient-to-r from-green-400 to-emerald-500 text-white  "
                    : "text-gray-600 hover:text-white hover:bg-gradient-to-r from-green-400 to-emerald-500"
                }`}
              >
                <div className="flex items-center space-x-3 focus:outline-none">
                  {item.icon}
                  <span>{item.label}</span>
                </div>
                {isDropdown && (
                  <>
                    {isOpen ? (
                      <ChevronDown className="w-4 h-4" />
                    ) : (
                      <ChevronRight className="w-4 h-4" />
                    )}
                  </>
                )}
              </button>

              {/* Submenu */}
              {isDropdown && isOpen && (
                <div className="ml-9 mt-1 space-y-1 outline-none">
                  {item.children!.map((child, i) => (
                    <button
                      key={i}
                      onClick={() => handleNavigate(child.path)}
                      className={`group relative flex items-center w-full px-4 py-3 rounded-xl text-sm font-medium bg-transparent hover:border-transparent  outline-none transition-all focus:outline-none ${
                        activePath === child.path
                          ? "hover:outline-none hover:ring-0 bg-gradient-to-r from-green-400 to-emerald-500 text-white "
                          : "text-gray-600 hover:text-white hover:bg-gradient-to-r from-green-400 to-emerald-500"
                      }`}
                    >
                      {child.label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </nav>
    </aside>
  );
}

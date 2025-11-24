import {
  Users,
  FileText,
  DollarSign,
  TrendingUp,
  ArrowUpRight,
} from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";

export default function Dashboard() {
  // Dummy data grafik
  const lineData = [
    { month: "Jan", value: 2400 },
    { month: "Feb", value: 1398 },
    { month: "Mar", value: 9800 },
    { month: "Apr", value: 3908 },
    { month: "May", value: 4800 },
    { month: "Jun", value: 3800 },
    { month: "Jul", value: 4300 },
  ];

  const pieData = [
    { name: "BPJS", value: 55 },
    { name: "General", value: 45 },
  ];
  const pieColors = ["#10B981", "#3B82F6"];

  const content = [
    // 📝 Reports
    <div className="ml-64 p-6 space-y-6">
      {/* Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mt-10">
        <div className="bg-white rounded-2xl shadow p-5 flex items-center justify-between hover:shadow-lg transition">
          <div>
            <p className="text-sm text-gray-500">Total Pasien</p>
            <h2 className="text-2xl font-bold text-gray-700">1,245</h2>
          </div>
          <div className="bg-green-100 text-green-600 p-3 rounded-xl">
            <Users className="w-6 h-6" />
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow p-5 flex items-center justify-between hover:shadow-lg transition">
          <div>
            <p className="text-sm text-gray-500">Tagihan Bulan Ini</p>
            <h2 className="text-2xl font-bold text-gray-700">Rp 85.4M</h2>
          </div>
          <div className="bg-blue-100 text-blue-600 p-3 rounded-xl">
            <FileText className="w-6 h-6" />
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow p-5 flex items-center justify-between hover:shadow-lg transition">
          <div>
            <p className="text-sm text-gray-500">Total Honor</p>
            <h2 className="text-2xl font-bold text-gray-700">Rp 22.1M</h2>
          </div>
          <div className="bg-yellow-100 text-yellow-600 p-3 rounded-xl">
            <DollarSign className="w-6 h-6" />
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow p-5 flex items-center justify-between hover:shadow-lg transition">
          <div>
            <p className="text-sm text-gray-500">Kinerja Bulan Ini</p>
            <div className="flex items-center gap-2">
              <h2 className="text-2xl font-bold text-gray-700">+12%</h2>
              <ArrowUpRight className="w-5 h-5 text-green-500" />
            </div>
          </div>
          <div className="bg-emerald-100 text-emerald-600 p-3 rounded-xl">
            <TrendingUp className="w-6 h-6" />
          </div>
        </div>
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Line Chart */}
        <div className="xl:col-span-2 bg-white p-6 rounded-2xl shadow hover:shadow-lg transition">
          <h2 className="text-lg font-semibold text-gray-700 mb-4">
            Tren Honor Per Bulan
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={lineData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#10B981"
                strokeWidth={3}
                dot={{ r: 4 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Pie Chart */}
        <div className="bg-white p-6 rounded-2xl shadow hover:shadow-lg transition">
          <h2 className="text-lg font-semibold text-gray-700 mb-4">
            Komposisi Pasien
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={pieData}
                dataKey="value"
                nameKey="name"
                cx="50%"
                cy="50%"
                outerRadius={100}
                label
              >
                {pieData.map((_entry, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={pieColors[index % pieColors.length]}
                  />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Activity / Summary Section */}
      <div className="bg-white rounded-2xl shadow p-6 hover:shadow-lg transition">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">
          Aktivitas Terbaru
        </h2>
        <ul className="divide-y divide-gray-100">
          <li className="py-3 flex justify-between">
            <span className="text-gray-600">
              🔹 Pasien <strong>BPJS</strong> menambah tagihan baru
            </span>
            <span className="text-gray-400 text-sm">2 jam lalu</span>
          </li>
          <li className="py-3 flex justify-between">
            <span className="text-gray-600">
              🔹 Honor dokter <strong>Dr. Ahmad</strong> telah diupdate
            </span>
            <span className="text-gray-400 text-sm">5 jam lalu</span>
          </li>
          <li className="py-3 flex justify-between">
            <span className="text-gray-600">
              🔹 Import data transaksi berhasil
            </span>
            <span className="text-gray-400 text-sm">Kemarin</span>
          </li>
        </ul>
      </div>
    </div>,
  ];

  return <div className="p-6">{content}</div>;
}

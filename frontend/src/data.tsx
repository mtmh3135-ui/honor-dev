import React, { useState } from "react";
import * as XLSX from "xlsx";

export default function XLSXPreview() {
  const [data, setData] = useState<string[][]>([]);

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (evt) => {
      const bstr = evt.target?.result;
      if (!bstr) return;

      const workbook = XLSX.read(bstr, { type: "binary" });
      const sheetName = workbook.SheetNames[0];
      const worksheet = workbook.Sheets[sheetName];
      const jsonData = XLSX.utils.sheet_to_json<string[]>(worksheet, { header: 1 });
      setData(jsonData);
    };
    reader.readAsBinaryString(file);
  };

  return (
    <div>
      <h2>Preview XLSX</h2>
      <input type="file" accept=".xlsx" onChange={handleFile} />
      <table border={1} style={{ borderCollapse: "collapse", marginTop: "10px" }}>
        <tbody>
          {data.slice(0, 20).map((row, i) => ( // tampilkan 20 baris pertama
            <tr key={i}>
              {row.map((cell, j) => (
                <td key={j} style={{ padding: "4px" }}>{cell}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {data.length > 20 && <p>... Total rows: {data.length}</p>}
    </div>
  );
}

import { useEffect, useState } from "react";
import reqprefix from "../utils";
import { ProxyRequest } from "../utils";

export default function Home() {
  const [data, setData] = useState<ProxyRequest[]>([]);

  useEffect(() => {
    const dataFetch = async () => {
      const resp = await fetch(reqprefix + "/api/proxy-requests");
      const parsedResp = await resp.json();
      if (parsedResp.error === null) {
        setData(parsedResp.data);
      } else {
        console.error(parsedResp.error);
      }
    };
    dataFetch();
  }, []);

  function handleDelete(id: string) {
    fetch(reqprefix + "/api/proxy-requests/{id}", { method: "DELETE" });
    setData(data.filter((r) => r.id != id));
  }

  return (
    <>
      <h1 className="text-center text-2xl mt-2 mb-8">Proxy Dashboard</h1>
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>Error</th>
            <th>ID</th>
            <th>Client IP</th>
            <th>Proxy Authorization</th>
            <th>Method</th>
            <th>URL</th>
            <th>Raw HTTP Request</th>
            <th>Raw HTTP Response</th>
            <th>Processing Time</th>
            <th>Upstream Response Time</th>
            <th>Delete</th>
          </tr>
        </thead>
        <tbody>
          {data?.map((req) => {
            return (
              <tr key={req.id}>
                <td>
                  {req.time == null
                    ? "Not logged."
                    : new Date(req.time * 1000).toLocaleString()}
                </td>
                <td>
                  {req.error == null ? (
                    "No errors."
                  ) : (
                    <span className="bg-red-50">{req.error}</span>
                  )}
                </td>
                <td>{req.id}</td>
                <td>{req.clientIP}</td>
                <td>{req.proxyAuthorization}</td>
                <td>{req.method}</td>
                <td title={req.url != null ? req.url : ""}>
                  {req.url != null && req.url.length > 60
                    ? req.url!.substring(0, 57) + "..."
                    : req.url}
                </td>
                <td
                  title={
                    req.rawHTTPRequest != null ? atob(req.rawHTTPRequest) : ""
                  }
                >
                  {req.rawHTTPRequest != null
                    ? atob(req.rawHTTPRequest).substring(0, 100) + "..."
                    : ""}
                </td>
                <td
                  title={
                    req.rawHTTPResponse != null ? atob(req.rawHTTPResponse) : ""
                  }
                >
                  {req.rawHTTPResponse != null
                    ? atob(req.rawHTTPResponse).substring(0, 100) + "..."
                    : ""}
                </td>
                <td>
                  {" "}
                  {req.processingTime != null
                    ? `${req.processingTime} ms`
                    : "Not logged."}
                </td>
                <td>
                  {req.upstreamResponseTime != null
                    ? `${req.upstreamResponseTime} ms`
                    : "Not logged."}
                </td>
                <td>
                  <button
                    onClick={() => {
                      handleDelete(req.id);
                    }}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}

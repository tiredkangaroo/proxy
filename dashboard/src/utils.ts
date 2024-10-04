const reqprefix = import.meta.env.DEV ? "http://localhost:1212" : "";

export interface ProxyRequest {
  clientIP: string | null;
  error: string | null;
  id: string;
  method: string | null;
  proxyAuthorization: string | null;
  rawHTTPRequest: string | null;
  rawHTTPResponse: string | null;
  time: number | null;
  processingTime: number | null;
  upstreamResponseTime: number | null;
  url: string | null;
}

export default reqprefix;

import { NextResponse } from "next/server";

// Short links are generated with this frontend's host (BASE_SHORT_URL on the
// API side), but the API still owns resolution — click tracking, expiry, and
// rate limiting all live there. This route just bounces the browser to the
// API's own /:code redirect so that logic isn't duplicated here.
const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ code: string }> },
) {
  const { code } = await params;
  return NextResponse.redirect(`${API_URL}/${code}`, 302);
}

import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";
import { getAccessToken } from "@/lib/session";

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  const accessToken = await getAccessToken();
  if (!accessToken) {
    return NextResponse.json({ error: "not authenticated" }, { status: 401 });
  }

  const { id } = await params;
  const body = await request.json();

  try {
    const link = await apiFetch(`/api/v1/links/${id}`, {
      method: "PATCH",
      accessToken,
      body: JSON.stringify(body),
    });
    return NextResponse.json(link);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "failed to update link" }, { status: 500 });
  }
}

export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const accessToken = await getAccessToken();
  if (!accessToken) {
    return NextResponse.json({ error: "not authenticated" }, { status: 401 });
  }

  const { id } = await params;

  try {
    await apiFetch(`/api/v1/links/${id}`, { method: "DELETE", accessToken });
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "failed to delete link" }, { status: 500 });
  }
}

export interface Link {
  id: string;
  shortCode: string;
  shortUrl: string;
  destination: string;
  expiresAt?: string;
  maxClicks?: number;
  clickCount: number;
  isActive: boolean;
  qrGenerated: boolean;
  createdAt: string;
}

export interface ClickEvent {
  id: string;
  linkId: string;
  clickedAt: string;
  referrer: string;
  userAgent: string;
}

export interface LinkDetail {
  link: Link;
  clicks: ClickEvent[];
}

import express from 'express';
import { inferEventType, resolveClientIp } from './auth-logging';

describe('auth logging helpers', () => {
  it('resolves client IP from forwarded headers', () => {
    const req = {
      headers: { 'x-forwarded-for': '198.51.100.1, 203.0.113.2' },
      ip: '10.0.0.1',
    } as unknown as express.Request;

    expect(resolveClientIp(req)).toBe('198.51.100.1');
  });

  it('falls back to request IP when no forwarded header is present', () => {
    const req = { headers: {}, ip: '10.0.0.5' } as unknown as express.Request;

    expect(resolveClientIp(req)).toBe('10.0.0.5');
  });

  it('maps auth paths to descriptive event types', () => {
    expect(inferEventType('/logout/keycloak')).toBe('sign-out');
    expect(inferEventType('/refresh/keycloak')).toBe('token-refresh');
    expect(inferEventType('/handler/frame')).toBe('sign-in-callback');
    expect(inferEventType('/failure')).toBe('sign-in-failure');
    expect(inferEventType('/start')).toBe('sign-in-start');
    expect(inferEventType('/other')).toBe('auth-request');
  });
});

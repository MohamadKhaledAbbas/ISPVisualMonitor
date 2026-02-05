import { describe, it, expect } from 'vitest';
import { formatBytes, formatBps, formatDuration, formatPercentage, getStatusColor, truncate } from '@/utils';

describe('formatBytes', () => {
  it('formats zero bytes', () => {
    expect(formatBytes(0)).toBe('0 Bytes');
  });

  it('formats bytes correctly', () => {
    expect(formatBytes(500)).toBe('500 Bytes');
  });

  it('formats kilobytes correctly', () => {
    expect(formatBytes(1024)).toBe('1 KB');
    expect(formatBytes(1536)).toBe('1.5 KB');
  });

  it('formats megabytes correctly', () => {
    expect(formatBytes(1048576)).toBe('1 MB');
    expect(formatBytes(1572864)).toBe('1.5 MB');
  });

  it('formats gigabytes correctly', () => {
    expect(formatBytes(1073741824)).toBe('1 GB');
  });
});

describe('formatBps', () => {
  it('formats zero bps', () => {
    expect(formatBps(0)).toBe('0 bps');
  });

  it('formats bps correctly', () => {
    expect(formatBps(500)).toBe('500 bps');
  });

  it('formats Kbps correctly', () => {
    expect(formatBps(1000)).toBe('1 Kbps');
  });

  it('formats Mbps correctly', () => {
    expect(formatBps(1000000)).toBe('1 Mbps');
  });

  it('formats Gbps correctly', () => {
    expect(formatBps(1000000000)).toBe('1 Gbps');
  });
});

describe('formatDuration', () => {
  it('formats minutes only', () => {
    expect(formatDuration(300)).toBe('5m');
  });

  it('formats hours and minutes', () => {
    expect(formatDuration(3660)).toBe('1h 1m');
  });

  it('formats days, hours, and minutes', () => {
    expect(formatDuration(90000)).toBe('1d 1h 0m');
  });
});

describe('formatPercentage', () => {
  it('formats with default decimals', () => {
    expect(formatPercentage(50.5)).toBe('50.5%');
  });

  it('formats with custom decimals', () => {
    expect(formatPercentage(50.567, 2)).toBe('50.57%');
  });
});

describe('getStatusColor', () => {
  it('returns green for active status', () => {
    expect(getStatusColor('active')).toBe('green');
    expect(getStatusColor('online')).toBe('green');
    expect(getStatusColor('up')).toBe('green');
  });

  it('returns red for inactive status', () => {
    expect(getStatusColor('inactive')).toBe('red');
    expect(getStatusColor('offline')).toBe('red');
    expect(getStatusColor('down')).toBe('red');
  });

  it('returns yellow for maintenance status', () => {
    expect(getStatusColor('maintenance')).toBe('yellow');
    expect(getStatusColor('degraded')).toBe('yellow');
  });

  it('returns gray for unknown status', () => {
    expect(getStatusColor('unknown')).toBe('gray');
  });
});

describe('truncate', () => {
  it('returns string as-is if shorter than length', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates string and adds ellipsis', () => {
    expect(truncate('hello world', 5)).toBe('hello...');
  });

  it('handles exact length', () => {
    expect(truncate('hello', 5)).toBe('hello');
  });
});

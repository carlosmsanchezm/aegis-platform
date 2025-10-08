import '@testing-library/jest-dom';

const suppressedErrorPatterns = [
  /Warning: findDOMNode is deprecated/,
  /Not implemented: HTMLCanvasElement\.prototype\.getContext/,
];

const suppressedWarnPatterns = [
  /React Router Future Flag Warning/, // v7 migration noise
];

const originalError = console.error;
// eslint-disable-next-line no-console
console.error = (...args: unknown[]) => {
  const firstArg = args[0];
  const message =
    typeof firstArg === 'string'
      ? firstArg
      : firstArg instanceof Error
      ? firstArg.message
      : '';
  if (suppressedErrorPatterns.some(pattern => pattern.test(message))) {
    return;
  }
  originalError(...args);
};

const originalWarn = console.warn;
// eslint-disable-next-line no-console
console.warn = (...args: unknown[]) => {
  const firstArg = args[0];
  const message = typeof firstArg === 'string' ? firstArg : '';
  if (suppressedWarnPatterns.some(pattern => pattern.test(message))) {
    return;
  }
  originalWarn(...args);
};

Object.defineProperty(window.HTMLCanvasElement.prototype, 'getContext', {
  value: () => null,
});

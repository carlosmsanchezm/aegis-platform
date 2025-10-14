import '@testing-library/jest-dom';

const suppressedErrorPatterns = [/Warning: findDOMNode is deprecated/];

const suppressedWarnPatterns = [/React Router Future Flag Warning/];

/* eslint-disable no-console */
const originalError = console.error;
console.error = (...args: unknown[]) => {
  const firstArg = args[0];
  let message = '';
  if (typeof firstArg === 'string') {
    message = firstArg;
  } else if (firstArg instanceof Error) {
    message = firstArg.message;
  }
  if (suppressedErrorPatterns.some(pattern => pattern.test(message))) {
    return;
  }
  originalError(...args);
};

const originalWarn = console.warn;
console.warn = (...args: unknown[]) => {
  const firstArg = args[0];
  const message = typeof firstArg === 'string' ? firstArg : '';
  if (suppressedWarnPatterns.some(pattern => pattern.test(message))) {
    return;
  }
  originalWarn(...args);
};
/* eslint-enable no-console */

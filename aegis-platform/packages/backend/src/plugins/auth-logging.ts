import express from 'express';
import {
  coreServices,
  createBackendModule,
} from '@backstage/backend-plugin-api';

const AUTH_BASE_PATH = '/api/auth';

export const resolveClientIp = (req: express.Request): string => {
  const forwarded = req.headers['x-forwarded-for'];
  if (Array.isArray(forwarded) && forwarded.length > 0) {
    return forwarded[0];
  }
  if (typeof forwarded === 'string' && forwarded.trim().length > 0) {
    return forwarded.split(',')[0]!.trim();
  }
  return req.ip;
};

export const inferEventType = (path: string): string => {
  const lowerPath = path.toLowerCase();
  if (lowerPath.includes('/logout')) {
    return 'sign-out';
  }
  if (lowerPath.includes('/refresh')) {
    return 'token-refresh';
  }
  if (lowerPath.includes('/handler') || lowerPath.includes('/frame')) {
    return 'sign-in-callback';
  }
  if (lowerPath.includes('/failure')) {
    return 'sign-in-failure';
  }
  if (lowerPath.includes('/start')) {
    return 'sign-in-start';
  }
  return 'auth-request';
};

export default createBackendModule({
  pluginId: 'auth',
  moduleId: 'audit-logging',
  register(env) {
    env.registerInit({
      deps: {
        logger: coreServices.logger,
        httpRouter: coreServices.httpRouter,
      },
      async init({ logger, httpRouter }) {
        const router = express.Router();

        router.use(AUTH_BASE_PATH, (req, res, next) => {
          const start = process.hrtime.bigint();
          const provider = req.path.split('/').filter(Boolean)[0] ?? 'unknown';
          const userAgent = req.get('user-agent') ?? 'unknown';
          const method = req.method.toUpperCase();
          const path = `${AUTH_BASE_PATH}${req.path}`;
          const eventType = inferEventType(req.path);
          const clientIp = resolveClientIp(req);

          res.on('finish', () => {
            const principal =
              (req.user as any)?.profile?.email ??
              (req.user as any)?.profile?.preferred_username ??
              (req.user as any)?.profile?.displayName ??
              'unresolved';
            const completed = new Date().toISOString();
            const durationMillis =
              Number(process.hrtime.bigint() - start) / 1_000_000;

            logger.info(
              `[auth-event] ts=${completed} event=${eventType} provider=${provider} method=${method} status=${
                res.statusCode
              } durationMs=${durationMillis.toFixed(
                2,
              )} path=${path} subject=${principal} ip=${clientIp} ua="${userAgent}"`,
            );
          });

          next();
        });

        httpRouter.use(router);
      },
    });
  },
});

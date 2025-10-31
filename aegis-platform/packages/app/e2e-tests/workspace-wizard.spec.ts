import { test, expect } from '@playwright/test';
import { failOnBrowserErrors } from '@backstage/e2e-test-utils/playwright';

failOnBrowserErrors();

test.describe('Workspace Wizard', () => {
  test('allows selecting a template, configuring, and launching', async ({ page }) => {
    await page.route('**/SubmitWorkload', route =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: 'w-e2e-123', status: 'PLACED', projectId: 'p-demo' }),
      }),
    );

    await page.goto('/aegis/workspaces/create');

    const signInButton = page.getByRole('button', { name: /sign in/i });
    if ((await signInButton.count()) > 0) {
      await signInButton.first().click();
    }

    await page.getByRole('button', { name: /python starter/i }).click();
    await page.getByRole('button', { name: /^next$/i }).click();

    await expect(page.getByLabel('Workspace ID')).toBeVisible();

    const workspaceId = page.getByLabel('Workspace ID');
    await workspaceId.fill('ws-e2e');

    await page.getByLabel('Storage (GiB)').fill('80');
    const ttlField = page.getByLabel('Auto shutdown (hours)');
    await ttlField.fill('6');

    const advancedToggle = page.getByRole('switch', { name: /advanced options/i });
    if (await advancedToggle.isEnabled()) {
      await advancedToggle.click();
    }

    const portsField = page.getByLabel('Exposed ports');
    await portsField.fill('22, 11111, 20000');

    const envField = page.getByLabel('Environment variables');
    await envField.fill('USER_NAME=aegis-test');

    await page.getByRole('button', { name: /^next$/i }).click();

    await expect(page.getByText(/Review & launch/i)).toBeVisible();

    await page.getByRole('button', { name: /launch workspace/i }).click();

    await expect(
      page.getByText(/Workspace w-e2e-123 was submitted/i),
    ).toBeVisible({ timeout: 5000 });
  });
});

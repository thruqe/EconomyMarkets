import puppeteer from 'puppeteer-core';
import fs from 'fs';
import path from 'path';

const outDir = path.resolve('/home/thruqe/.gemini/antigravity/brain/388073ed-62be-43ce-bf5c-8fd4af5b719e/screenshots');
if (!fs.existsSync(outDir)) {
  fs.mkdirSync(outDir, { recursive: true });
}

async function clickTab(page: any, label: string) {
  console.log(`Navigating to tab: ${label}...`);
  await page.evaluate((tabLabel: string) => {
    const allButtons = Array.from(document.querySelectorAll('header button, header div button'));
    let direct = allButtons.find(b => b.textContent?.trim().toLowerCase() === tabLabel.toLowerCase());
    if (direct) {
      direct.click();
      return;
    }

    // Try opening More dropdown if available
    const moreBtn = allButtons.find(b => b.textContent?.toLowerCase().includes('more'));
    if (moreBtn && window.innerWidth >= 1024 && window.innerWidth < 1280) {
      moreBtn.click();
    }

    // Check if on mobile drawer
    const menuBtn = document.querySelector('header button[aria-label="Toggle Navigation Menu"]') as HTMLElement;
    const drawerOpen = document.querySelector('header .z-50');
    if (menuBtn && window.innerWidth < 1024 && !drawerOpen) {
      menuBtn.click();
    }
  }, label);

  await new Promise(r => setTimeout(r, 400));

  await page.evaluate((tabLabel: string) => {
    const allButtons = Array.from(document.querySelectorAll('header button, header div[class*="z-50"] button, header div.absolute button'));
    const target = allButtons.find(b => b.textContent?.toLowerCase().includes(tabLabel.toLowerCase())) as HTMLElement;
    if (target) {
      target.click();
    }
  }, label);

  await new Promise(r => setTimeout(r, 800));
}

async function clickUniverseSubTab(page: any, label: string) {
  console.log(`Clicking Universe sub-tab: ${label}...`);
  await page.evaluate((tabLabel: string) => {
    const buttons = Array.from(document.querySelectorAll('main button'));
    const target = buttons.find(b => b.textContent?.toLowerCase().includes(tabLabel.toLowerCase())) as HTMLElement;
    if (target) target.click();
  }, label);
  await new Promise(r => setTimeout(r, 600));
}

async function run() {
  const browser = await puppeteer.launch({
    executablePath: '/usr/bin/google-chrome',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
    headless: true,
  });

  const page = await browser.newPage();
  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', err => console.error('BROWSER ERROR:', err));

  // --- 1. Full Desktop (1440x900) ---
  console.log('--- 1. FULL DESKTOP (1440x900) ---');
  await page.setViewport({ width: 1440, height: 900 });
  await page.goto('http://localhost:8080', { waitUntil: 'networkidle0', timeout: 15000 });
  await new Promise(r => setTimeout(r, 1200));
  await page.screenshot({ path: path.join(outDir, '01_overview_desktop_1440.png') });

  // Universe - Domestic
  await clickTab(page, 'Universe');
  await clickUniverseSubTab(page, 'Domestic Stocks');
  await page.screenshot({ path: path.join(outDir, '02_universe_domestic_1440.png') });

  // Universe - World Stocks
  await clickUniverseSubTab(page, 'World Stocks');
  await page.screenshot({ path: path.join(outDir, '03_universe_world_1440.png') });

  // Universe - Sovereign Bonds
  await clickUniverseSubTab(page, 'Sovereign Bonds');
  await page.screenshot({ path: path.join(outDir, '04_universe_bonds_1440.png') });

  // Universe - Forex & Currencies
  await clickUniverseSubTab(page, 'Forex & Currencies');
  await page.screenshot({ path: path.join(outDir, '05_universe_forex_1440.png') });

  // Nation Builder - Maintenance, Tourism, TFP, and Loan Requests
  await clickTab(page, 'Nation');
  await page.screenshot({ path: path.join(outDir, '06_nation_maintenance_tourism_1440.png') });

  // Macro Dashboard - Yield Curve & Tourism receipts
  await clickTab(page, 'Macro');
  await page.screenshot({ path: path.join(outDir, '08_macro_yield_curve_1440.png') });

  // Terminal with TradingView Lightweight Chart
  await clickTab(page, 'Terminal');
  await new Promise(r => setTimeout(r, 1200));
  await page.screenshot({ path: path.join(outDir, '09_terminal_tradingview_1440.png') });

  // Policy Studio
  await clickTab(page, 'Policy');
  await page.screenshot({ path: path.join(outDir, '10_policy_studio_1440.png') });

  // --- 2. Intermediate Desktop (1100px) ---
  console.log('--- 2. INTERMEDIATE DESKTOP (1100px) ---');
  await page.setViewport({ width: 1100, height: 800 });
  await clickTab(page, 'Universe');
  await clickUniverseSubTab(page, 'Sovereign Bonds');
  await page.screenshot({ path: path.join(outDir, '11_reduced_viewport_1100.png') });

  // --- 3. Tablet Landscape (1024x768) ---
  console.log('--- 3. TABLET LANDSCAPE (1024x768) ---');
  await page.setViewport({ width: 1024, height: 768 });
  await clickTab(page, 'Nation');
  await page.screenshot({ path: path.join(outDir, '12_tablet_landscape_1024.png') });

  // --- 4. Tablet Portrait (768x1024) ---
  console.log('--- 4. TABLET PORTRAIT (768x1024) ---');
  await page.setViewport({ width: 768, height: 1024 });
  await clickTab(page, 'Overview');
  await page.screenshot({ path: path.join(outDir, '13_tablet_portrait_768.png') });

  // --- 5. Mobile (390x844) ---
  console.log('--- 5. MOBILE (390x844) ---');
  await page.setViewport({ width: 390, height: 844 });
  await new Promise(r => setTimeout(r, 600));
  await page.screenshot({ path: path.join(outDir, '14_mobile_390.png') });

  await browser.close();
  console.log('Puppeteer multi-asset & responsive suite completed successfully!');
}

run().catch(console.error);

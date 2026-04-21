# DESIGN.md

> 把 Kafka 子页从“说明过多的展示面板”收回到“轻量、高密度、好扫读的运维后台”。

## 1. Visual Theme & Atmosphere

**Style**: Lean Console
**Keywords**: 克制、轻量、运维感、结构清晰、低装饰、高信息密度、稳定、直接
**Tone**: 管理后台、审计工具、运维工作台 — NOT 营销页、作品集、内容型长文页面
**Feel**: 像一张整理过的控制台，不抢戏，只让操作者更快看到重点。

**Interaction Tier**: L1 精致静态
**Dependencies**: CSS only

## 2. Color Palette & Roles

```css
:root {
  /* Backgrounds */
  --bg: #eef3f8;
  --surface: #ffffff;
  --surface-alt: #f8fbff;
  --surface-hover: #f3f7fc;

  /* Borders */
  --border: rgba(148, 163, 184, 0.18);
  --border-hover: rgba(47, 125, 246, 0.28);

  /* Text */
  --text: #0f172a;
  --text-secondary: #475569;
  --text-tertiary: #64748b;

  /* Accent */
  --accent: #2f7df6;
  --accent-hover: #1f6be3;

  /* RGB variants for rgba() */
  --bg-rgb: 238, 243, 248;
  --accent-rgb: 47, 125, 246;

  /* Semantic */
  --success: #16a34a;
  --error: #dc2626;
  --warning: #d97706;
}
```

**Color Rules:**
- 所有页面颜色统一走 CSS 变量，禁止在子页组件里新增零散 hex。
- 风险色只用于状态和局部提示，不做大面积铺底。
- 同一块面板内只允许一个强调点，避免同时出现“高亮标题 + 彩色背景 + 彩色按钮”。

## 3. Typography Rules

**Font Stack:**
```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap');
```

| Role | Font | Size | Weight | Line Height | Letter Spacing |
|------|------|------|--------|-------------|----------------|
| Hero H1 | var(--font-display) | 2rem | 800 | 1.05 | -0.04em |
| Section H2 | var(--font-display) | 1.125rem | 700 | 1.2 | -0.02em |
| H3 | var(--font-display) | 0.95rem | 700 | 1.25 | -0.01em |
| Body | var(--font-sans) | 0.875rem | 400 | 1.55 | 0 |
| Label | var(--font-sans) | 0.75rem | 600 | 1.4 | 0.02em |
| Mono/Code | var(--font-mono) | 0.8125rem | 500 | 1.5 | 0 |

**Typography Rules:**
- 子页主标题只保留一句短描述，长度控制在 18-26 个汉字。
- 卡片说明文默认不超过两行，超过就改成更短的标签式信息。
- 快捷操作、风险摘要、统计文案优先“短词 + 数值”，不要写成长句。
- **NEVER use**: 超大号中文标题、三行以上副文案、按钮里塞入解释性句子。

**Text Decoration:**
- Hero h1: 无渐变、无投影，只保留高对比粗体。
- Section h2 / h3: 无渐变、无投影。

## 4. Component Stylings

### Buttons
```css
.el-button {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 12px;
}

.el-button--primary {
  background: var(--accent);
  border-color: var(--accent);
}

.el-button--primary:hover {
  background: var(--accent-hover);
  border-color: var(--accent-hover);
}

.el-button.is-quiet {
  background: var(--surface-alt);
  border: 1px solid var(--border);
  color: var(--text-secondary);
}
```

### Cards
```css
.content-card,
.workspace-panel,
.page-metric-card {
  border: 1px solid var(--border);
  border-radius: 20px;
  background: var(--surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.workspace-panel {
  padding: 16px 18px;
}

.page-metric-card {
  min-height: 96px;
  padding: 14px 16px;
}
```

### Navigation
```css
.page-eyebrow {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
```

### Links
```css
a,
.el-button.is-link {
  color: var(--accent);
}

a:hover,
.el-button.is-link:hover {
  color: var(--accent-hover);
}
```

### Tags / Badges
```css
.el-tag {
  min-height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 0.6875rem;
  font-weight: 600;
}
```

### Quick Action Filters
```css
.quick-filter-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.quick-filter-btn {
  min-height: 42px;
  font-size: 0.8125rem;
  font-weight: 600;
}
```

## 5. Layout Principles

**Container:**
- Max width: 100%
- Padding: 16-24px
- Narrow variant (text-heavy): 760px

**Spacing Scale:**
- Section padding: 16-18px
- Component gap: 10-14px
- Card internal padding: 14-18px

**Grid:**
```css
.workbench-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.page-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 12px;
}
```

## 6. Depth & Elevation

| Level | Treatment | Use |
|-------|-----------|-----|
| Flat | 无阴影，仅边框 | 过滤栏、次级容器 |
| Subtle | `0 10px 24px rgba(15, 23, 42, 0.05)` | 主卡片、统计卡片 |
| Elevated | `0 14px 28px rgba(15, 23, 42, 0.08)` | Hover 态、抽屉、弹窗 |

## 7. Animation & Interaction

**Motion Philosophy**: 只做帮助理解结构的轻交互，不做会增加认知负担的装饰动画。
**Tier**: L1

### Dependencies
```html
无额外动画依赖
```

### Base Setup
```js
无
```

### Entrance Animation
```css
.content-card,
.workspace-panel,
.page-metric-card {
  transition: border-color 0.18s ease, box-shadow 0.18s ease, background-color 0.18s ease;
}
```

### Scroll Behavior
```js
无滚动联动动效
```

### Hover & Focus States
```css
.workspace-panel:hover,
.page-metric-card:hover {
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.08);
}

.quick-filter-btn:hover {
  border-color: var(--border-hover);
}
```

### Special Effects
- 不使用视差、不使用滚动驱动、不使用大面积背景动画。
- 风险模块只保留颜色和选中态，不叠加额外动效。

### Reduced Motion
```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation: none !important;
    transition-duration: 0.01ms !important;
  }
}
```

## 8. Do's and Don'ts

### Do
- 用更短的模块标题和一句话副标题。
- 让统计卡片优先展示数值，其次才是说明。
- 把复杂说明收敛到抽屉、弹窗、详情页，而不是首页模块。
- 快捷操作优先使用短动词短语。
- 保持同一页的卡片密度一致。

### Don't
- ❌ 不要在每个子页头部写两三句解释性长文案。
- ❌ 不要让一个快捷入口同时承载说明、状态、风险、动作四层信息。
- ❌ 不要把按钮做成营销卡片。
- ❌ 不要在小屏上保留三行以上的副说明。
- ❌ 不要给每个子页都塞满“建议”“提示”“摘要”模块。
- ❌ 不要重复表达相同信息，例如标题说一次、卡片再说一次、表头再说一次。
- ❌ 不要把操作区字体做得比主表格还大。
- ❌ 不要让页面一屏内出现太多彩色背景块。

## 9. Responsive Behavior

**Breakpoints:**
| Name | Width | Key Changes |
|------|-------|-------------|
| Desktop | > 1120px | 双列工作区、三列快捷筛选 |
| Tablet | 641px-1120px | 工作区单列、快捷筛选两列 |
| Mobile | < 640px | 所有面板单列、快捷筛选单列 |

**Touch Targets:** minimum 44px
**Collapsing Strategy:** 标题、副文案、标签优先压缩，内容表格和关键操作保留

```css
@media (max-width: 1120px) {
  .workbench-grid {
    grid-template-columns: 1fr;
  }

  .quick-filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .page-container {
    padding: 14px;
  }

  .page-header h2 {
    font-size: 1.75rem;
  }

  .quick-filter-grid {
    grid-template-columns: 1fr;
  }
}
```

# Research Report: Odoo ERP System Architecture

**Date:** 2026-02-19
**Version focus:** Odoo 16/17 (Community & Enterprise)
**Sources:** Odoo official docs (training knowledge), GitHub odoo/odoo source analysis, developer references

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Module/Plugin System](#1-moduleplugin-system)
3. [Core Architecture](#2-core-architecture)
4. [ORM / Data Layer](#3-orm--data-layer)
5. [Event System](#4-event-system)
6. [View / UI System](#5-view--ui-system)
7. [Inter-module Communication Patterns](#6-inter-module-communication-patterns)
8. [Key Architectural Patterns](#7-key-architectural-patterns)
9. [Unresolved Questions](#unresolved-questions)

---

## Executive Summary

Odoo is a monolithic-but-modular Python/JavaScript ERP framework. The server runs on Python (Flask-like WSGI via Werkzeug), PostgreSQL backend, and an ORM that maps Python classes to DB tables. Modules ("addons") are Python packages with a mandatory `__manifest__.py` that declare dependencies; the framework topologically sorts them at startup. The ORM uses metaclass magic to merge inherited models at class-definition time. The frontend (Odoo 17) is an OWL (Odoo Web Library) SPA — a custom reactive component framework replacing the legacy Widget system. Modules extend UI by registering components/actions into a central registry, not by subclassing. Inter-module comms use Python method override chains (via `super()`), XML-based view inheritance (XPath), and a server-side bus/channel system for real-time events.

---

## 1. Module/Plugin System

### 1.1 Addon Structure

Every module is a directory (Python package) on the addon path:

```
my_module/
├── __init__.py          # Python package entry
├── __manifest__.py      # Module metadata & dependencies
├── models/
│   ├── __init__.py
│   └── my_model.py
├── views/
│   └── my_view.xml
├── security/
│   └── ir.model.access.csv
├── data/
│   └── demo_data.xml
├── static/
│   └── src/
│       ├── js/          # OWL components (Odoo 17)
│       └── xml/         # QWeb templates
└── controllers/
    └── main.py          # HTTP routes
```

### 1.2 Manifest File (`__manifest__.py`)

```python
{
    'name': 'My Module',
    'version': '17.0.1.0.0',
    'category': 'Accounting',
    'summary': 'Short description',
    'description': '...',
    'author': 'My Company',
    'website': 'https://...',
    'license': 'LGPL-3',
    'depends': ['base', 'mail', 'account'],   # Dependency list (topological order enforced)
    'data': [                                  # XML/CSV loaded in order
        'security/ir.model.access.csv',
        'views/my_view.xml',
        'data/data.xml',
    ],
    'demo': ['data/demo.xml'],
    'assets': {                                # JS/CSS asset bundles (Odoo 16+)
        'web.assets_backend': [
            'my_module/static/src/js/my_component.js',
            'my_module/static/src/xml/my_template.xml',
        ],
    },
    'installable': True,
    'auto_install': False,                     # Auto-install if all depends installed
    'application': True,                       # Shows in Apps menu
}
```

### 1.3 Module Loading Sequence

1. **Addon path scan** — `odoo.conf` specifies `addons_path` directories; framework scans for `__manifest__.py`
2. **Dependency resolution** — DAG topological sort on `depends` graph; circular deps = error at startup
3. **Module graph** — `odoo.modules.graph.Graph` builds the load order
4. **Python import** — `__init__.py` of each module imported in dependency order; models auto-discovered via import
5. **Registry build** — `odoo.modules.registry.Registry` merges all model classes
6. **Data loading** — XML/CSV data files loaded in manifest `data` list order per module

### 1.4 Dependency Rules

- `depends: ['base']` is implicit minimum; `base` always loads first
- `auto_install: True` + `depends: ['sale', 'purchase']` = installs automatically when both deps present
- **External dependencies** declared in `external_dependencies: {'python': ['lxml'], 'bin': ['wkhtmltopdf']}`
- At runtime, `env['ir.module.module']` tracks install state in DB; upgrade/uninstall triggers re-migration

---

## 2. Core Architecture

### 2.1 Server Stack

```
HTTP Request
    │
    ▼
Werkzeug WSGI
    │
    ▼
odoo.http.Root (dispatcher)
    │
    ├── Static file serving
    │
    └── odoo.http.Application
            │
            ▼
        Route matching (@http.route decorators)
            │
            ▼
        Controller method
            │
            ▼
        Environment (env) creation
        [Registry + Cursor + User]
            │
            ▼
        ORM / Business Logic
            │
            ▼
        PostgreSQL (psycopg2)
```

### 2.2 The Environment (`env`)

The `Environment` object is the central context object passed everywhere:

```python
env = request.env          # From HTTP context
env = self.env             # From model methods

env.user                   # res.users record of current user
env.uid                    # User ID integer
env.company                # Current company
env.cr                     # DB cursor (psycopg2)
env.context                # Dict: lang, tz, active_id, etc.

# Accessing models:
env['sale.order']          # Returns model class bound to this env
env['sale.order'].search([('state', '=', 'sale')])
```

The `Environment` is immutable per request but can be altered via `env.with_user()`, `env.with_context()`, `env.with_company()`.

### 2.3 Registry

`odoo.modules.registry.Registry` is a per-database singleton that:
- Stores all model classes merged from all installed addons
- Rebuilds on module install/uninstall/upgrade
- Thread-safe via read/write locks
- Maps `'model.name'` → merged Python class

### 2.4 HTTP Controllers

```python
from odoo import http

class MyController(http.Controller):
    @http.route('/my/path', auth='user', type='json')
    def my_endpoint(self, **kw):
        return {'result': 'ok'}

    @http.route('/web/download', auth='public', type='http')
    def download(self):
        return http.request.make_response(data, headers=[...])
```

Routes register via class-level decorator scanning; `auth` options: `'user'`, `'public'`, `'none'`.

---

## 3. ORM / Data Layer

### 3.1 Model Definition

```python
from odoo import models, fields, api

class SaleOrder(models.Model):
    _name = 'sale.order'          # DB table: sale_order
    _description = 'Sales Order'
    _inherit = ['mail.thread', 'mail.activity.mixin']  # Mixins
    _order = 'date_order desc'
    _rec_name = 'name'

    name = fields.Char(required=True, copy=False, default='New')
    date_order = fields.Datetime(default=fields.Datetime.now)
    partner_id = fields.Many2one('res.partner', required=True)
    order_line = fields.One2many('sale.order.line', 'order_id')
    amount_total = fields.Monetary(compute='_compute_amount', store=True)
    state = fields.Selection([
        ('draft', 'Quotation'),
        ('sale', 'Sales Order'),
        ('cancel', 'Cancelled'),
    ], default='draft')
```

### 3.2 Field Types

| Field | SQL type | Notes |
|-------|----------|-------|
| `Char` | varchar | `size` limit optional |
| `Text` | text | Multiline |
| `Integer` | int4 | |
| `Float` | float8 | `digits=(precision, scale)` |
| `Monetary` | float8 | Uses `currency_id` field |
| `Boolean` | bool | |
| `Date` | date | |
| `Datetime` | timestamp | UTC stored |
| `Binary` | bytea / filestore | `attachment=True` for filesystem |
| `Selection` | varchar | Static list or callable |
| `Many2one` | int4 FK | `ondelete='cascade/set null/restrict'` |
| `One2many` | virtual | Inverse of Many2one, not stored |
| `Many2many` | relation table | Auto-creates join table |
| `Html` | text | Sanitized HTML |
| `Json` | jsonb (Odoo 17) | |

### 3.3 Computed Fields

```python
@api.depends('order_line.price_total')
def _compute_amount(self):
    for order in self:
        order.amount_total = sum(order.order_line.mapped('price_total'))
```

- `store=True` → persisted to DB, recomputed on dependency change
- `store=False` (default) → computed on-the-fly each read
- `@api.depends_context('lang')` → recompute when context changes

### 3.4 Model Inheritance

Three inheritance modes:

**1. Classical (extension) — most common:**
```python
class ResPartner(models.Model):
    _inherit = 'res.partner'     # Extends existing model IN PLACE
    my_field = fields.Char()     # Adds column to existing res_partner table
```

**2. Prototype (copy):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = 'res.partner'     # Copies fields/methods; new table created
```

**3. Delegation (SQL inheritance):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherits = {'res.partner': 'partner_id'}  # FK delegation; fields proxied
    partner_id = fields.Many2one('res.partner', required=True, ondelete='cascade')
```

### 3.5 ORM Internals — Metaclass Magic

All model classes use `ModelMetaclass` which:
1. At import time, registers class in `odoo.models._model_registry` (module-level dict)
2. At registry build time, merges all `_inherit` classes via Python MRO
3. Final class is a true Python class combining all modules' contributions
4. Field objects are descriptors on the merged class

### 3.6 Domain Syntax (Query DSL)

```python
# Odoo domain = list of conditions (Polish notation AND/OR)
domain = [
    ('state', '=', 'sale'),
    ('date_order', '>=', '2024-01-01'),
    '|',
    ('partner_id.country_id.code', '=', 'US'),
    ('partner_id.country_id.code', '=', 'CA'),
]
records = env['sale.order'].search(domain, limit=100, order='date_order desc')
```

Operators: `=`, `!=`, `>`, `>=`, `<`, `<=`, `in`, `not in`, `like`, `ilike`, `=like`, `=ilike`, `child_of`, `parent_of`.

### 3.7 Constraints

```python
# SQL constraint (DB-level)
_sql_constraints = [
    ('name_unique', 'UNIQUE(name, company_id)', 'Order name must be unique per company'),
]

# Python constraint (ORM-level)
@api.constrains('date_start', 'date_end')
def _check_dates(self):
    for rec in self:
        if rec.date_start > rec.date_end:
            raise ValidationError("Start must be before end")
```

---

## 4. Event System

Odoo has multiple "event-like" mechanisms — no single unified event bus for server-side model communication.

### 4.1 ORM Method Hooks (Primary Pattern)

Override `create`, `write`, `unlink` on any model:

```python
@api.model_create_multi
def create(self, vals_list):
    records = super().create(vals_list)    # Call parent chain
    for rec in records:
        rec._do_something_on_create()
    return records

def write(self, vals):
    result = super().write(vals)
    if 'state' in vals:
        self._on_state_change()
    return result
```

The `super()` chain propagates through all `_inherit` extensions in MRO order.

### 4.2 Onchange (UI-only Events)

Triggered client-side when field changes, calls server:

```python
@api.onchange('partner_id')
def _onchange_partner(self):
    if self.partner_id:
        self.payment_term_id = self.partner_id.property_payment_term_id
        return {'domain': {'child_ids': [('parent_id', '=', self.partner_id.id)]}}
```

Note: `onchange` does NOT fire on `create`/`write` — only in UI form view.

### 4.3 Automated Actions (`ir.actions.server`)

DB-stored rules that fire on model events:
- Trigger: `on_create`, `on_write`, `on_create_or_write`, `on_unlink`, `on_change`, `on_time`
- Actions: run Python code, send email, create record, update fields, call webhook

### 4.4 Mail Chatter / Tracking

`mail.thread` mixin provides field tracking:

```python
class MyModel(models.Model):
    _inherit = ['mail.thread']

    name = fields.Char(tracking=True)       # Logs changes to chatter
    state = fields.Selection(tracking=10)   # tracking=priority (lower = more visible)
```

### 4.5 Real-time Bus (Longpolling / WebSocket)

`bus.bus` model for server-push notifications:

```python
# Server sends message to channel
self.env['bus.bus']._sendone(channel, message_type, payload)
self.env['bus.bus']._sendmany([
    (channel1, 'type1', payload1),
    (channel2, 'type2', payload2),
])

# Channels are strings or tuples: e.g., ('res.partner', partner_id)
# Client subscribes via JS bus service
```

Odoo 16+ uses WebSocket (`bus.websocket`) instead of longpolling.

### 4.6 Scheduled Actions (Cron)

```python
# In XML:
<record id="ir_cron_my_job" model="ir.cron">
    <field name="name">My Daily Job</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="state">code</field>
    <field name="code">model._my_cron_method()</field>
    <field name="interval_type">days</field>
    <field name="interval_number">1</field>
</record>
```

---

## 5. View / UI System

### 5.1 Architecture Overview (Odoo 17)

```
Browser
  │
  └── OWL SPA (Single Page Application)
        │
        ├── WebClient (root component)
        │     ├── ActionManager
        │     │     └── Views (List, Form, Kanban, ...)
        │     ├── NavBar
        │     ├── HomeMenu
        │     └── Services (rpc, notification, dialog, ...)
        │
        └── OWL Component Tree (reactive, signal-based)
```

### 5.2 OWL Framework (Odoo Web Library)

Odoo's custom reactive UI framework (NOT React/Vue):

```js
// Component definition (Odoo 17)
import { Component, useState, xml } from "@odoo/owl";

class MyWidget extends Component {
    static template = xml`
        <div t-on-click="onClick">
            <t t-esc="state.count"/>
        </div>
    `;

    setup() {
        this.state = useState({ count: 0 });
    }

    onClick() {
        this.state.count++;
    }
}
```

Features: signals/reactive state, QWeb templates (XML-based), lifecycle hooks, slots, sub-templates.

### 5.3 Server-Side Views (XML)

Views are stored in DB (`ir.ui.view`) and fetched via RPC:

```xml
<!-- Form view -->
<record id="view_sale_order_form" model="ir.ui.view">
    <field name="name">sale.order.form</field>
    <field name="model">sale.order</field>
    <field name="arch" type="xml">
        <form>
            <header>
                <button name="action_confirm" string="Confirm" type="object" class="btn-primary"/>
                <field name="state" widget="statusbar"/>
            </header>
            <sheet>
                <field name="name"/>
                <field name="partner_id"/>
                <notebook>
                    <page string="Order Lines">
                        <field name="order_line">
                            <tree>
                                <field name="product_id"/>
                                <field name="qty"/>
                                <field name="price_unit"/>
                            </tree>
                        </field>
                    </page>
                </notebook>
            </sheet>
            <div class="oe_chatter">
                <field name="message_follower_ids"/>
                <field name="message_ids"/>
            </div>
        </form>
    </field>
</record>
```

**View types:** `form`, `list` (formerly `tree`), `kanban`, `search`, `calendar`, `gantt`, `pivot`, `graph`, `activity`, `map`, `cohort`

### 5.4 View Inheritance (XPath)

Modules extend views without modifying originals:

```xml
<record id="view_sale_order_form_inherit_my_module" model="ir.ui.view">
    <field name="name">sale.order.form.inherit.my_module</field>
    <field name="model">sale.order</field>
    <field name="inherit_id" ref="sale.view_sale_order_form"/>
    <field name="arch" type="xml">
        <!-- XPath to locate injection point -->
        <xpath expr="//field[@name='partner_id']" position="after">
            <field name="my_custom_field"/>
        </xpath>

        <!-- Shorthand: field-based targeting -->
        <field name="amount_total" position="replace">
            <field name="amount_total" widget="monetary_custom"/>
        </field>
    </field>
</record>
```

Positions: `before`, `after`, `inside`, `replace`, `attributes`.

### 5.5 Frontend Module System (Odoo 16/17)

JavaScript uses ES modules with Odoo's custom bundler (not Webpack/Vite):

```js
// Registering a custom field widget
import { registry } from "@web/core/registry";
import { Component } from "@odoo/owl";

class MyCustomWidget extends Component { ... }

registry.category("fields").add("my_widget", {
    component: MyCustomWidget,
    supportedTypes: ["char"],
});
```

**Central Registry categories:**
- `fields` — custom field widgets
- `views` — custom view types
- `actions` — action handlers
- `services` — singleton services (rpc, notification, user, ...)
- `systray` — top-bar icons
- `client_actions` — full-page client-side actions

### 5.6 QWeb Templates

Server-side (Python, for PDF/email) and client-side (JS):

```xml
<!-- Client-side QWeb in OWL -->
<t t-name="my_module.MyTemplate">
    <div>
        <t t-foreach="items" t-as="item" t-key="item.id">
            <span t-esc="item.name"/>
        </t>
        <t t-if="showButton">
            <button t-on-click="() => this.doAction()">Click</button>
        </t>
    </div>
</t>
```

Directives: `t-if`, `t-else`, `t-elif`, `t-foreach`, `t-as`, `t-key`, `t-esc`, `t-raw`, `t-call`, `t-set`, `t-att-*`, `t-on-*`

### 5.7 RPC Communication Pattern

```js
// From OWL component / service
import { useService } from "@web/core/utils/hooks";

setup() {
    this.orm = useService("orm");
    this.rpc = useService("rpc");
}

async fetchData() {
    // High-level ORM calls
    const records = await this.orm.searchRead(
        "sale.order",
        [["state", "=", "sale"]],
        ["name", "partner_id", "amount_total"],
        { limit: 10 }
    );

    // Low-level JSON-RPC
    const result = await this.rpc("/web/dataset/call_kw", {
        model: "sale.order",
        method: "my_custom_method",
        args: [],
        kwargs: { context: {} },
    });
}
```

---

## 6. Inter-module Communication Patterns

### 6.1 Python Method Resolution Order (MRO)

When multiple modules inherit the same model, Python MRO determines call order:

```
Module C inherits Module B inherits Module A (base)
Call order: C.method → B.method → A.method (via super())
```

Each extension MUST call `super()` to maintain the chain.

### 6.2 Signal Pattern via `_inherit`

No explicit event bus for most server-side model events. Pattern:

1. Module A defines `sale.order` with `state` field
2. Module B (accounting) does `_inherit = 'sale.order'` and overrides `action_confirm`
3. When sale confirmed, accounting logic automatically runs via override chain

### 6.3 Cross-model References

```python
# Via related fields
invoice_ids = fields.Many2many('account.move', compute='_compute_invoices')

# Via SQL
self.env.cr.execute("SELECT ... FROM ...")

# Via model search
invoices = self.env['account.move'].search([('invoice_origin', '=', self.name)])
```

---

## 7. Key Architectural Patterns

### 7.1 What Makes Odoo Unique

| Pattern | Odoo approach |
|---------|---------------|
| Module extension | Python class merging via metaclass at registry build |
| View extension | XPath injection into XML stored in DB |
| JS extension | Central registry pattern (no subclassing) |
| DB schema | Auto-migrated; ORM drives `ALTER TABLE` on install/upgrade |
| Multi-tenancy | Per-DB registries; single process serves multiple DBs |
| Security | `ir.model.access` (CRUD per model per group) + `ir.rule` (record-level domain filters) |

### 7.2 Migration System

```python
# migrations/17.0.1.1.0/pre-migrate.py
def migrate(cr, version):
    cr.execute("ALTER TABLE sale_order ADD COLUMN my_col varchar")

# migrations/17.0.1.1.0/post-migrate.py
def migrate(cr, version):
    env = api.Environment(cr, SUPERUSER_ID, {})
    env['sale.order'].search([]).write({'my_col': 'default'})
```

### 7.3 Multi-company Architecture

- `company_id` field on most models
- `ir.rule` records filter by `env.company_ids`
- `company_dependent=True` fields store per-company values in `ir.property`

---

## Unresolved Questions

1. **Odoo 17 vs 16 asset bundling**: The new bundler in 17.0 differs from the old `ir.qweb` bundle system — exact caching/invalidation behavior needs verification against actual source.
2. **WebSocket bus internals**: Exact channel subscription protocol in `bus.websocket` (Odoo 16/17) — how workers coordinate not fully confirmed.
3. **OWL signals vs useState**: Odoo 17 introduced fine-grained reactivity signals (`@odoo/owl` v2) — exact API stability needs checking against latest commits.
4. **Enterprise vs Community delta**: Several UI components (Gantt, Map, Studio) are Enterprise-only; their extension APIs may differ.
5. **Thread safety in multi-worker (gevent) mode**: How registry rebuild interacts with concurrent requests under gevent workers needs source-level confirmation.

---

*Report generated from training knowledge (Odoo 16/17, cutoff Aug 2025). No live source fetching was possible in this session due to tool restrictions.*

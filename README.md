<a id="readme-top"></a>

<br />
<div align="center">

  <!-- Hero Header -->
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./docs/assets/hero-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="./docs/assets/hero-light.svg">
    <img alt="ecos - Open FinOps Data Stack" src="./docs/assets/hero-light.svg" width="100%">
  </picture>

  <br />

  <h3 align="center">
    > <a href="https://ecos-labs.io/">ecos-labs.io</a>
  </h3>
  <p align="center">
    <a href="https://github.com/ecos-labs/ecos/issues">Issues</a>
    ·
    <a href="https://github.com/ecos-labs/ecos/blob/main/CONTRIBUTING.md">Contribution</a>
  </p>
</div>

---

- [What is ecos?](#what-is-ecos)
- [What's included?](#whats-included)
- [Quickstart](#quickstart)
- [Contributing](#contributing)

## What is ecos?

**ecos** is an open source FinOps data stack that transforms AWS Cost and Usage Reports (CUR) into clean, enriched, high-performance datasets. Its analytics-ready semantic layer enables cost transparency, allocation, and optimization with actionable insights.

**Highlights:**

- **Own Your Data** - Runs in your infrastructure, any cloud or warehouse. Full transform transparency, data never leaves your account.

- **Modular & Extensible** - 40+ pre-built data models. Start small, extend with your business logic and data, sources unified automatically.

- **Production Ready** - Smart data materialization balances speed and cost. Fast CLI deployment, serverless scaling, full dbt extensibility.

- **Advanced Analytics** - Semantic layer for BI and AI agents. Pinpoint cost drivers, expose savings potentials, track trends and more.

<p align="center">
  <em>AWS Cost and Usage Reports (CUR) with Athena, today. All clouds, eventually.</em>
</p>

## What's included?

- **ecos CLI** - One-command project setup, provisioning, and model deployment
- **dbt Models** - 40+ pre-built SQL models for cost analysis and optimization
  - **Bronze** - Views: raw CUR data from S3
  - **Silver** - Views/incremental tables: cleaned, normalized, mapped data
  - **Gold** - Incremental Tables (partitioned): business-ready analytics (pre-computed, fast queries)
  - **Serve** - Custom views you create based on your needs
- **Web UI** - Interactive [lineage graph](https://ecos-labs.io/lineage) and [data catalog](https://ecos-labs.io/data) browser
- **MCP Server** - *(Preview)* AI-powered cost insights via Model Context Protocol

## Quickstart

**For detailed setup** view the full [Quickstart](https://ecos-labs.io/docs/cli-quickstart) guide.

**1. Enable AWS CUR**
Setup AWS Cost & Usage Reports for Athena.

**2. Install ecos CLI**
```bash
# Install via brew
brew tap ecos-labs/homebrew-ecos
brew install ecos

# Verify installation
ecos version
```

**3. Initialize Project**
```bash
mkdir ecos-playground && cd ecos-playground
ecos init
```

**4. Transform Your Data**
```bash
# Load reference data
ecos transform seed

# Run all transformations
ecos transform run
```

**5. Query & Analyze**

```sql
-- Daily service costs
select
  usage_date,
  service_name,
  sum(total_effective_cost) as cost
from gold_core__service_daily
where billing_period = '2025-10'
group by 1, 2
order by 3 desc;
```

## Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, improving documentation, or helping others. Your contributions are greatly appreciated.

- **[Issues](https://github.com/ecos-labs/ecos/issues)** - Report features or bugs
- **[Contributing Guide](https://github.com/ecos-labs/ecos/blob/main/CONTRIBUTING.md)** - Learn how to contribute
- **[Code of Conduct](https://github.com/ecos-labs/ecos/blob/main/CODE_OF_CONDUCT.md)** - Our community standards

---

<div align="center">
  <sub>Built with ❤️ by <a href="https://ecos-labs.io/about">ecos-labs.io</a></sub>
</div>

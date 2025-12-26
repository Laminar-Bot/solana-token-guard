# Skill Matrix: Which Skill to Use When

This document helps you quickly identify which skill(s) to invoke for different scenarios.

## Quick Reference Table

| Scenario | Primary Skill | Supporting Skills |
|----------|--------------|-------------------|
| **Code Review** | Snarky Senior Engineer | Security Sentinel, QA Engineer |
| **System Design** | Pragmatic Architect | API Platform, Data Engineer, Security Sentinel |
| **New Feature Planning** | Product Engineering Lead | Pragmatic Architect, DevEx Champion |
| **Security Audit** | Security Sentinel | Compliance Guardian, API Platform |
| **Production Incident** | Incident Commander | SRE, Observability Engineer |
| **Performance Optimization** | SRE | Pragmatic Architect, Observability Engineer |
| **Cost Reduction** | FinOps Optimizer | Pragmatic Architect, ML Pragmatist |
| **Developer Tooling** | DevEx Champion | Platform Builder |
| **Board Presentation** | Executive Liaison | Pragmatic Architect, Product Lead |
| **Team Culture Issue** | Empathetic Team Lead | Executive Liaison, Staff+ IC Advisor |
| **ML/AI Decision** | ML Pragmatist | Pragmatic Architect, Data Engineer |
| **Compliance/Audit** | Compliance Guardian | Security Sentinel, Technical Writer |
| **API Design** | API Platform Engineer | Pragmatic Architect, Security Sentinel |
| **Data Pipeline** | Data Engineer | ML Pragmatist, Observability Engineer |
| **Mobile App** | Mobile Platform Engineer | Frontend/UX Specialist, QA Engineer |
| **Legacy Refactoring** | Legacy Archaeologist | Snarky Senior Engineer, QA Engineer |
| **Monitoring Setup** | Observability Engineer | SRE, Platform Builder |
| **Internal Platform** | Platform Builder | DevEx Champion, Pragmatic Architect |
| **Developer Community** | Developer Advocate | DevEx Champion, Technical Writer, Open Source Strategist |
| **Enterprise Sales POC** | Solutions Architect | Pragmatic Architect, Product Lead, Security Sentinel |
| **Staff+ IC Promotion** | Staff+ IC Advisor | Empathetic Team Lead, Pragmatic Architect |
| **Open Source Strategy** | Open Source Strategist | Developer Advocate, Security Sentinel, Compliance Guardian |
| **M&A Acquisition Evaluation** | M&A Due Diligence Specialist | Pragmatic Architect, Security Sentinel, Technical Recruiting |
| **Vendor Consolidation** | Vendor Management Strategist | FinOps Optimizer, Security Sentinel, Platform Builder |
| **Hiring Pipeline Scaling** | Technical Recruiting Strategist | Executive Liaison, Empathetic Team Lead, DevEx Champion |
| **Agile/Org Transformation** | Engineering Transformation Leader | Empathetic Team Lead, DevEx Champion, Executive Liaison |
| **AI Bias Detection** | AI Ethics & Governance Officer | ML Pragmatist, Compliance Guardian, Technical Recruiting |
| **Data Governance Setup** | Data Strategy Officer (CDO) | Pragmatic Architect, Compliance Guardian, Data Engineer |
| **Design System Creation** | UI Design System Architect | Visual Design Specialist, Product Designer, Frontend/UX Specialist |
| **User Research Study** | UX Research & Strategy Lead | Product Designer, Product Engineering Lead |
| **UI/UX Redesign** | Product Designer | UX Research Lead, Visual Design, Interaction Design, Frontend/UX |
| **Brand Identity Design** | Visual Design & Brand Specialist | Product Designer, UI Design System Architect |
| **Animation & Motion Design** | Motion Design & Animation Engineer | Interaction Design Specialist, Frontend/UX Specialist |
| **Micro-Interaction Design** | Interaction Design Specialist | Motion Design Engineer, Frontend/UX Specialist |
| **Content Strategy & SEO** | Content Strategist / Technical Marketing | Developer Advocate, Technical Writer, Product Designer |
| **Accessibility Compliance** | Accessibility Specialist | Frontend/UX Specialist, Product Designer, Compliance Guardian |
| **Internationalization (i18n)** | Localization & i18n Engineer | Frontend/UX Specialist, Data Engineer, Product Designer |
| **Product Growth & Analytics** | Growth Engineer / Product Analytics | Data Engineer, Product Engineering Lead, UX Research Lead |
| **Infrastructure as Code** | DevOps / IaC Specialist | Cloud Architect, SRE, Platform Builder |
| **Microservices Architecture** | Backend / Distributed Systems Engineer | Pragmatic Architect, API Platform, SRE |
| **Privacy & GDPR Compliance** | Privacy Engineer | Compliance Guardian, Security Sentinel, Data Strategy Officer |
| **Enterprise B2B Integration** | Enterprise Integration Architect | Solutions Architect, API Platform, Data Engineer |

---

## Decision Trees

### "Should I build this feature?"

```
Start → Product Engineering Lead (business value?)
      ↓
      → Pragmatic Architect (how to build?)
      ↓
      → Security Sentinel (is it secure?)
      ↓
      → FinOps Optimizer (what's the cost?)
      ↓
      → DevEx Champion (dev impact?)
      ↓
      → Executive Liaison (present to leadership)
```

### "System is down!"

```
INCIDENT → Incident Commander (lead response)
         ↓
         → SRE (mitigate/restore)
         ↓
         → Observability Engineer (investigate)
         ↓
         → Executive Liaison (stakeholder comms)
         ↓
         → Post-Mortem (all hands)
```

### "We need to comply with GDPR"

```
GDPR → Compliance Guardian (requirements)
     ↓
     → Pragmatic Architect (data architecture)
     ↓
     → Security Sentinel (encryption, access control)
     ↓
     → Data Engineer (data deletion workflows)
     ↓
     → Technical Writer (documentation)
```

### "Our ML model is underperforming"

```
ML Issue → ML Pragmatist (model evaluation)
         ↓
         → Data Engineer (data quality check)
         ↓
         → Observability Engineer (monitor drift)
         ↓
         → FinOps Optimizer (cost vs accuracy trade-off)
```

### "Building a developer community"

```
Community → Developer Advocate (community strategy)
          ↓
          → DevEx Champion (docs & DX check)
          ↓
          → Technical Writer (content creation)
          ↓
          → Open Source Strategist (OSS components)
          ↓
          → Product Lead (business metrics)
```

### "Enterprise customer wants POC"

```
POC Request → Solutions Architect (scoping & tech validation)
            ↓
            → Product Lead (deal size & ROI check)
            ↓
            → Pragmatic Architect (architecture design)
            ↓
            → Platform Builder (reusable components)
            ↓
            → Executive Liaison (stakeholder comms)
```

### "Senior engineer wants Staff promotion"

```
Career Path → Staff+ IC Advisor (expectations & roadmap)
            ↓
            → Empathetic Team Lead (timeline & feedback)
            ↓
            → Pragmatic Architect (assign domain ownership)
            ↓
            → Executive Liaison (executive sponsorship)
            ↓
            → DevEx Champion (high-visibility projects)
```

### "Should we open source this project?"

```
OSS Decision → Open Source Strategist (strategy & licensing)
             ↓
             → Security Sentinel (code audit)
             ↓
             → Compliance Guardian (legal review)
             ↓
             → Developer Advocate (maintenance commitment)
             ↓
             → DevEx Champion (brand impact)
```

### "Redesigning our product UX"

```
UX Redesign → UX Research Lead (user needs & validation)
            ↓
            → Product Designer (journey maps & wireframes)
            ↓
            → Visual Design Specialist (visual hierarchy & brand)
            ↓
            → Interaction Design Specialist (micro-interactions & states)
            ↓
            → Motion Design Engineer (animations & transitions)
            ↓
            → UI Design System Architect (component library updates)
            ↓
            → Frontend/UX Specialist (implementation & a11y)
```

---

## By Problem Domain

### Architecture & Design

**When:** Designing new systems, evaluating trade-offs, making technical decisions

**Skills:**
1. **Pragmatic Architect** - Overall system design
2. **API Platform Engineer** - API contracts and versioning
3. **Data Engineer** - Data modeling and pipelines
4. **Security Sentinel** - Security architecture
5. **FinOps Optimizer** - Cost modeling

### Code Quality & Maintenance

**When:** Code reviews, refactoring, tech debt

**Skills:**
1. **Snarky Senior Engineer** - Code quality and patterns
2. **Legacy Archaeologist** - Refactoring legacy code
3. **QA Automation Engineer** - Test coverage
4. **Technical Writer** - Documentation

### Operations & Reliability

**When:** Production issues, monitoring, scaling

**Skills:**
1. **SRE** - Reliability and uptime
2. **Incident Commander** - Crisis management
3. **Observability Engineer** - Monitoring and debugging
4. **FinOps Optimizer** - Infrastructure cost

### Product & Business

**When:** Feature planning, roadmap, stakeholder management

**Skills:**
1. **Product Engineering Lead** - Feature prioritization
2. **Executive Liaison** - Board/CEO communication
3. **Pragmatic Architect** - Technical feasibility
4. **FinOps Optimizer** - ROI analysis

### Team & Culture

**When:** Hiring, team dynamics, career development

**Skills:**
1. **Empathetic Team Lead** - People management
2. **DevEx Champion** - Developer happiness
3. **Executive Liaison** - Managing up
4. **Technical Writer** - Knowledge sharing

### Specialized Domains

**When:** Domain-specific challenges

**Skills:**
1. **ML Pragmatist** - AI/ML decisions
2. **Compliance Guardian** - Regulatory requirements
3. **Mobile Platform Engineer** - Mobile apps
4. **Platform Builder** - Internal tooling
5. **Frontend/UX Specialist** - User interfaces
6. **API Platform Engineer** - API design

### Design & User Experience

**When:** UI/UX design, research, branding, design systems

**Skills:**
1. **Product Designer** - End-to-end product design and user flows
2. **UX Research & Strategy Lead** - User research and data-driven insights
3. **UI Design System Architect** - Design systems and component libraries
4. **Visual Design & Brand Specialist** - Visual design, typography, brand identity
5. **Interaction Design Specialist** - Micro-interactions and behavioral design
6. **Motion Design & Animation Engineer** - Animations and motion design
7. **Frontend/UX Specialist** - Implementation and accessibility

### Developer Relations & Growth

**When:** Community building, enterprise sales, open source, career development

**Skills:**
1. **Developer Advocate** - Community engagement and developer experience
2. **Solutions Architect** - Enterprise POCs and customer onboarding
3. **Staff+ IC Advisor** - Senior IC career mentorship
4. **Open Source Strategist** - OSS strategy and governance

### Strategic Operations & C-Level

**When:** M&A, vendor management, recruiting at scale, org transformation, AI ethics, data strategy

**Skills:**
1. **M&A Due Diligence Specialist** - Acquisition evaluation and integration
2. **Vendor Management Strategist** - SaaS optimization and contract negotiation
3. **Technical Recruiting Strategist** - Hiring systems and talent pipeline
4. **Engineering Transformation Leader** - Org redesign and culture change
5. **AI Ethics & Governance Officer** - Responsible AI and bias mitigation
6. **Data Strategy Officer (CDO)** - Data governance and analytics platforms

### Content, Marketing & Growth

**When:** Content strategy, SEO, technical marketing, product growth, analytics, experimentation

**Skills:**
1. **Content Strategist / Technical Marketing** - Content strategy, SEO, developer-focused storytelling
2. **Growth Engineer / Product Analytics** - A/B testing, metrics, growth loops, retention analysis
3. **Developer Advocate** - Community content and technical evangelism
4. **Technical Writer** - Documentation and clarity

### Accessibility & Localization

**When:** WCAG compliance, assistive technology, internationalization, global expansion

**Skills:**
1. **Accessibility Specialist** - WCAG compliance, screen readers, keyboard navigation, ARIA
2. **Localization & i18n Engineer** - i18n architecture, RTL support, locale formatting
3. **Frontend/UX Specialist** - Implementation of a11y and i18n features
4. **Compliance Guardian** - Legal compliance (ADA, Section 508, EAA)

### Infrastructure & DevOps

**When:** Infrastructure automation, GitOps, immutable infrastructure, drift detection

**Skills:**
1. **DevOps / IaC Specialist** - Terraform, GitOps, immutable infrastructure, secrets management
2. **Cloud Architect** - Multi-cloud strategy and infrastructure design
3. **SRE** - Reliability and production operations
4. **Platform Builder** - Self-service infrastructure and golden paths

### Backend & Distributed Systems

**When:** Microservices architecture, event-driven systems, service mesh, distributed transactions

**Skills:**
1. **Backend / Distributed Systems Engineer** - Microservices, event-driven architecture, saga patterns, service mesh
2. **Pragmatic Architect** - System design and service boundaries
3. **API Platform Engineer** - API contracts and versioning
4. **SRE** - Operational resilience and observability

### Privacy & Data Protection

**When:** GDPR/CCPA compliance, consent management, data minimization, DSARs

**Skills:**
1. **Privacy Engineer** - Privacy by design, GDPR/CCPA implementation, consent systems, DSAR automation
2. **Compliance Guardian** - Regulatory frameworks and legal compliance
3. **Security Sentinel** - Data security and encryption
4. **Data Strategy Officer** - Data governance and retention policies

### Enterprise Integration

**When:** B2B integrations, connecting to Salesforce/SAP/NetSuite, iPaaS, webhooks

**Skills:**
1. **Enterprise Integration Architect** - iPaaS, ESB, enterprise connectors, API gateways
2. **Solutions Architect** - Pre-sales POCs and enterprise customer integration
3. **API Platform Engineer** - API design and versioning for external partners
4. **Data Engineer** - ETL/ELT for data warehouse integration

---

## Multi-Skill Workflows

### New Product Launch

**Phase 1: Planning**
- Product Engineering Lead (define MVP)
- Pragmatic Architect (high-level design)
- FinOps Optimizer (budget)

**Phase 2: Design**
- Pragmatic Architect (detailed architecture)
- API Platform Engineer (API contracts)
- Data Engineer (data model)
- Security Sentinel (threat model)
- Compliance Guardian (regulatory check)

**Phase 3: Design**
- UX Research Lead (user validation)
- Product Designer (flows and wireframes)
- Visual Design Specialist (visual design)
- Interaction Design Specialist (interactions)
- UI Design System Architect (component specs)

**Phase 4: Build**
- Snarky Senior Engineer (code review)
- QA Automation Engineer (test plan)
- DevEx Champion (tooling)
- Frontend/UX Specialist (UI implementation)
- Motion Design Engineer (animations)
- Mobile Platform Engineer (if mobile)

**Phase 5: Deploy**
- SRE (deployment strategy)
- Observability Engineer (monitoring)
- Platform Builder (CI/CD)

**Phase 6: Launch**
- Incident Commander (on standby)
- Executive Liaison (comms)

**Phase 7: Post-Launch**
- Observability Engineer (metrics)
- Product Engineering Lead (feedback)
- UX Research Lead (user feedback analysis)
- FinOps Optimizer (cost review)

### Security Incident Response

1. **Incident Commander** - Coordinate response
2. **Security Sentinel** - Assess breach scope
3. **SRE** - Contain and isolate
4. **Observability Engineer** - Forensics
5. **Compliance Guardian** - Notification requirements
6. **Executive Liaison** - Stakeholder communication
7. **Technical Writer** - Post-mortem documentation

### Platform Migration

1. **Pragmatic Architect** - Migration strategy
2. **Legacy Archaeologist** - Assess current state
3. **Platform Builder** - Build new platform
4. **Data Engineer** - Data migration plan
5. **Security Sentinel** - Security audit
6. **SRE** - Rollout plan
7. **DevEx Champion** - Developer training
8. **Technical Writer** - Migration docs

---

## When to Use the Skill Orchestrator

**Use Skill Orchestrator when:**
- You need input from 3+ personas
- There's no clear primary skill
- You need to resolve conflicting viewpoints
- The problem is cross-cutting (architecture + security + cost + people)

**Examples:**
- "Should we rewrite our monolith in microservices?"
- "How should we approach this RFP from an enterprise client?"
- "What's our strategy for scaling from 10 to 100 engineers?"

The Orchestrator will convene a "roundtable" and synthesize recommendations.

---

## Skill Collaboration Patterns

### Security Sentinel + API Platform
**Collaboration:** API security reviews (auth, rate limiting, input validation)

### DevEx Champion + Platform Builder
**Collaboration:** Internal developer platform features

### ML Pragmatist + Data Engineer
**Collaboration:** ML pipeline design (data quality, feature engineering, model training)

### Compliance Guardian + Security Sentinel
**Collaboration:** Regulatory security requirements (GDPR, HIPAA, SOC2)

### SRE + Observability Engineer
**Collaboration:** Production monitoring and incident response

### Pragmatic Architect + FinOps Optimizer
**Collaboration:** Cost-aware architecture decisions

### Product Engineering Lead + Empathetic Team Lead
**Collaboration:** Balancing delivery pressure with team health

### Executive Liaison + Pragmatic Architect
**Collaboration:** Translating technical decisions to business stakeholders

### Developer Advocate + Open Source Strategist
**Collaboration:** Building and maintaining OSS community projects

### Solutions Architect + Platform Builder
**Collaboration:** Creating reusable enterprise integration patterns

### Staff+ IC Advisor + Empathetic Team Lead
**Collaboration:** Career development for senior engineers

### Open Source Strategist + Compliance Guardian
**Collaboration:** OSS licensing and legal compliance

### Developer Advocate + DevEx Champion
**Collaboration:** Developer community tooling and documentation

### Solutions Architect + Executive Liaison
**Collaboration:** Strategic customer engagement and contract support

### Product Designer + UX Research Lead
**Collaboration:** User-centered design (research insights → design decisions)

### UI Design System Architect + Frontend/UX Specialist
**Collaboration:** Design system implementation and component development

### Visual Design Specialist + Interaction Design Specialist
**Collaboration:** Visual polish and behavioral design cohesion

### Motion Design Engineer + Frontend/UX Specialist
**Collaboration:** Performance-optimized animation implementation

### Product Designer + Product Engineering Lead
**Collaboration:** Balancing user needs with technical and business constraints

### Content Strategist + Developer Advocate
**Collaboration:** Developer-focused content creation and distribution strategy

### Growth Engineer + UX Research Lead
**Collaboration:** Data-driven product optimization (quantitative metrics + qualitative insights)

### Accessibility Specialist + Frontend/UX Specialist
**Collaboration:** WCAG-compliant UI implementation and assistive technology testing

### Localization Engineer + Product Designer
**Collaboration:** Culturally-adapted UX for global markets (RTL, locale-specific patterns)

### DevOps/IaC Specialist + Cloud Architect
**Collaboration:** Infrastructure automation and multi-cloud GitOps workflows

### Backend/Distributed Systems Engineer + SRE
**Collaboration:** Resilient microservices architecture with observability and circuit breakers

### Privacy Engineer + Compliance Guardian
**Collaboration:** GDPR/CCPA compliance implementation with legal frameworks and privacy-by-design architecture

### Enterprise Integration Architect + Solutions Architect
**Collaboration:** Pre-sales POCs with enterprise system integrations (Salesforce, NetSuite, SAP)

---

## Tips for Multi-Skill Sessions

1. **Start with the Orchestrator** if you're unsure which skills to use
2. **Use sequential consultation** when outputs depend on each other
3. **Use parallel consultation** when gathering independent perspectives
4. **Document decisions** - especially cross-skill agreements
5. **Revisit with new context** - skills may give different advice as situation evolves

---

## Skill Update Frequency

This matrix should be reviewed and updated:
- **Quarterly:** As new skills are added
- **After major incidents:** If skill gaps are identified
- **When patterns emerge:** If certain skill combinations are frequently used

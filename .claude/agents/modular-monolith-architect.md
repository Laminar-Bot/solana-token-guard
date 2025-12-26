---
name: modular-monolith-architect
description: Enforces modular monolith architecture boundaries and patterns for CryptoRugMunch. Use when creating new modules, validating imports, refactoring code, or checking for circular dependencies.
tools: Read, Edit, Grep, Bash
model: sonnet
skills: pragmatic-architect, snarky-senior-engineer
---

# Modular Monolith Architecture Enforcer

You enforce strict module boundaries and architectural standards for CryptoRugMunch's modular monolith.

## Architecture Rules (STRICT - NO EXCEPTIONS)

### Module Structure (Standard Template)

```
src/modules/{module-name}/
├── {module}.service.ts       # Business logic (pure, testable)
├── {module}.repository.ts    # Database access (Prisma only)
├── {module}.controller.ts    # API routes (Fastify)
├── {module}.types.ts         # TypeScript types/interfaces
├── {module}.errors.ts        # Module-specific errors
├── index.ts                  # PUBLIC API (only exports - other modules import from here!)
└── __tests__/
    ├── {module}.service.test.ts
    ├── {module}.repository.test.ts
    └── {module}.integration.test.ts
```

### Import Rules (ENFORCED BY HOOKS + ESLINT)

✅ **ALLOWED** (Public API imports):
```typescript
// Other modules can ONLY import from index.ts
import { ScanService } from '@/modules/scan';
import { UserRepository } from '@/modules/user';
import type { Scan, ScanResult } from '@/modules/scan';
```

❌ **FORBIDDEN** (Direct internal imports - will be blocked):
```typescript
// Direct imports to internal files bypass the public API
import { ScanService } from '@/modules/scan/scan.service'; // ❌ NO!
import { calculateScore } from '@/modules/scan/helpers'; // ❌ NO!
import type { MetricResult } from '@/modules/scan/types'; // ❌ NO!
```

**Why this rule exists**:
- Public API (`index.ts`) acts as a contract
- Internal refactoring doesn't break consumers
- Clear separation of public vs. internal implementation
- Easier to extract modules into microservices later

### Dependency Rules (No Circular Dependencies)

```
ALLOWED Dependencies (acyclic):

scan → blockchain-api ✅
scan → user ✅
telegram → scan ✅
telegram → user ✅
payment → user ✅
payment → subscription ✅

FORBIDDEN (creates cycles):

user → scan ❌
blockchain-api → scan ❌
scan → telegram ❌
```

**Check for cycles**:
```bash
npx madge --circular --extensions ts src/modules/
```

### Module Template Generator

```typescript
// scripts/new-module.ts
import { writeFile, mkdir } from 'fs/promises';
import path from 'path';

export async function generateModule(moduleName: string) {
  const pascalCase = toPascalCase(moduleName); // 'scan' → 'Scan'
  const kebabCase = toKebabCase(moduleName);   // 'tokenAnalysis' → 'token-analysis'

  const basePath = `src/modules/${kebabCase}`;

  // Create directory structure
  await mkdir(basePath, { recursive: true });
  await mkdir(`${basePath}/__tests__`, { recursive: true });

  // Generate service.ts
  await writeFile(`${basePath}/${kebabCase}.service.ts`, `
import { ${pascalCase}Repository } from './${kebabCase}.repository';
import type { Create${pascalCase}Input, ${pascalCase} } from './${kebabCase}.types';
import { ${pascalCase}Error } from './${kebabCase}.errors';
import { logger } from '@/shared/logger';

export class ${pascalCase}Service {
  constructor(private repository: ${pascalCase}Repository) {}

  async create(data: Create${pascalCase}Input): Promise<${pascalCase}> {
    try {
      logger.info({ data }, 'Creating ${moduleName}');
      const result = await this.repository.create(data);
      logger.info({ id: result.id }, '${pascalCase} created successfully');
      return result;
    } catch (error) {
      logger.error({ error, data }, 'Failed to create ${moduleName}');
      throw new ${pascalCase}Error('Failed to create ${moduleName}', { cause: error });
    }
  }

  async findById(id: string): Promise<${pascalCase} | null> {
    return await this.repository.findById(id);
  }

  async findMany(): Promise<${pascalCase}[]> {
    return await this.repository.findMany();
  }

  async update(id: string, data: Partial<Create${pascalCase}Input>): Promise<${pascalCase}> {
    try {
      logger.info({ id, data }, 'Updating ${moduleName}');
      const result = await this.repository.update(id, data);
      logger.info({ id }, '${pascalCase} updated successfully');
      return result;
    } catch (error) {
      logger.error({ error, id, data }, 'Failed to update ${moduleName}');
      throw new ${pascalCase}Error('Failed to update ${moduleName}', { cause: error });
    }
  }

  async delete(id: string): Promise<void> {
    try {
      logger.info({ id }, 'Deleting ${moduleName}');
      await this.repository.delete(id);
      logger.info({ id }, '${pascalCase} deleted successfully');
    } catch (error) {
      logger.error({ error, id }, 'Failed to delete ${moduleName}');
      throw new ${pascalCase}Error('Failed to delete ${moduleName}', { cause: error });
    }
  }
}
  `.trim());

  // Generate repository.ts
  await writeFile(`${basePath}/${kebabCase}.repository.ts`, `
import { prisma } from '@/shared/database';
import type { Create${pascalCase}Input, ${pascalCase} } from './${kebabCase}.types';

export class ${pascalCase}Repository {
  async create(data: Create${pascalCase}Input): Promise<${pascalCase}> {
    return await prisma.${moduleName}.create({ data });
  }

  async findById(id: string): Promise<${pascalCase} | null> {
    return await prisma.${moduleName}.findUnique({ where: { id } });
  }

  async findMany(): Promise<${pascalCase}[]> {
    return await prisma.${moduleName}.findMany();
  }

  async update(id: string, data: Partial<Create${pascalCase}Input>): Promise<${pascalCase}> {
    return await prisma.${moduleName}.update({
      where: { id },
      data
    });
  }

  async delete(id: string): Promise<void> {
    await prisma.${moduleName}.delete({ where: { id } });
  }
}
  `.trim());

  // Generate types.ts
  await writeFile(`${basePath}/${kebabCase}.types.ts`, `
export interface ${pascalCase} {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  // Add your fields here
}

export interface Create${pascalCase}Input {
  // Add your input fields here
}

export interface Update${pascalCase}Input {
  // Add your update fields here
}
  `.trim());

  // Generate errors.ts
  await writeFile(`${basePath}/${kebabCase}.errors.ts`, `
export class ${pascalCase}Error extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = '${pascalCase}Error';
  }
}
  `.trim());

  // Generate index.ts (PUBLIC API)
  await writeFile(`${basePath}/index.ts`, `
/**
 * ${pascalCase} Module - Public API
 *
 * Only exports listed here can be imported by other modules.
 * Internal implementation details are encapsulated.
 */

export { ${pascalCase}Service } from './${kebabCase}.service';
export { ${pascalCase}Repository } from './${kebabCase}.repository';
export { ${pascalCase}Error } from './${kebabCase}.errors';
export type {
  ${pascalCase},
  Create${pascalCase}Input,
  Update${pascalCase}Input
} from './${kebabCase}.types';
  `.trim());

  // Generate test template
  await writeFile(`${basePath}/__tests__/${kebabCase}.service.test.ts`, `
import { describe, it, expect, beforeEach } from 'vitest';
import { ${pascalCase}Service } from '../${kebabCase}.service';
import { ${pascalCase}Repository } from '../${kebabCase}.repository';

describe('${pascalCase}Service', () => {
  let service: ${pascalCase}Service;
  let repository: ${pascalCase}Repository;

  beforeEach(() => {
    repository = new ${pascalCase}Repository();
    service = new ${pascalCase}Service(repository);
  });

  describe('create', () => {
    it('should create a new ${moduleName}', async () => {
      // TODO: Implement test
      expect(true).toBe(true);
    });
  });

  describe('findById', () => {
    it('should find ${moduleName} by id', async () => {
      // TODO: Implement test
      expect(true).toBe(true);
    });
  });
});
  `.trim());

  console.log(`✅ Module '${moduleName}' generated successfully!`);
  console.log(`\nFiles created:`);
  console.log(`  ${basePath}/${kebabCase}.service.ts`);
  console.log(`  ${basePath}/${kebabCase}.repository.ts`);
  console.log(`  ${basePath}/${kebabCase}.controller.ts`);
  console.log(`  ${basePath}/${kebabCase}.types.ts`);
  console.log(`  ${basePath}/${kebabCase}.errors.ts`);
  console.log(`  ${basePath}/index.ts (PUBLIC API)`);
  console.log(`  ${basePath}/__tests__/${kebabCase}.service.test.ts`);
  console.log(`\nNext steps:`);
  console.log(`  1. Update Prisma schema: prisma/schema.prisma`);
  console.log(`  2. Run migration: npx prisma migrate dev --name add_${kebabCase}`);
  console.log(`  3. Implement business logic in: ${basePath}/${kebabCase}.service.ts`);
  console.log(`  4. Write tests in: ${basePath}/__tests__/`);
}

function toPascalCase(str: string): string {
  return str
    .split(/[-_]/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join('');
}

function toKebabCase(str: string): string {
  return str
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .toLowerCase();
}
```

## Validation Commands

### `/validate-boundaries`
```bash
# Check all imports follow public API rules
npx eslint src/modules --rule 'no-restricted-imports: error'

# Or use madge for visual dependency graph
npx madge --circular --extensions ts --image deps.svg src/modules/
```

### `/dependency-graph`
```bash
# Visualize module dependencies (check for cycles)
npx madge --circular --extensions ts src/modules/

# If cycles found, it will list them:
# ✖ Found 1 circular dependency!
# 1) user -> scan -> user
```

### `/new-module <name>`
```bash
# Generate new module with full template
npm run new:module -- token-analysis

# Creates:
# src/modules/token-analysis/
# ├── token-analysis.service.ts
# ├── token-analysis.repository.ts
# ├── token-analysis.types.ts
# ├── token-analysis.errors.ts
# ├── index.ts
# └── __tests__/
```

## ESLint Configuration (Enforce Boundaries)

```javascript
// .eslintrc.js
module.exports = {
  rules: {
    'no-restricted-imports': ['error', {
      patterns: [
        {
          group: ['../*/*.service', '../*/*.repository', '../*/*.controller'],
          message: 'Import from module public API (index.ts) only. Direct imports to internal files are forbidden.'
        }
      ]
    }]
  }
};
```

## Related Documentation

- `docs/03-TECHNICAL/architecture/modular-monolith-structure.md` - Architecture specification
- `docs/03-TECHNICAL/architecture/system-architecture.md` - Overall system design
- `.claude/api-integration-rules.md` - Import rules and standards
- Fastify docs (Context7): `/fastify/fastify` - Route patterns

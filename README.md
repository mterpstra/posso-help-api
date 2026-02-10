# ZapManejo API

Go backend for cattle ranch management. Receives WhatsApp messages, parses them into structured data, and stores in MongoDB. Also provides REST API for the dashboard.

## WhatsApp Message Formats

The API accepts WhatsApp messages in specific formats. Messages are parsed line by line, and multiple entries can be sent in a single message.

### Birth Messages

Records a new animal birth in the herd.

**Format:**
```
{tag} {sex} {breed}
{area}
```

**Fields:**
- `tag` - Numeric ear tag number (required, must be > 0)
- `sex` - `m` or `f` (case insensitive)
- `breed` - Must match a known breed or account-specific breed nickname
- `area` - Optional, on a separate line. If not recognized as existing area, creates a new one
- `date` - Optional, format `dd/mm` on any line

**Default Breeds:**
angus, nelore, brangus, sta.zelia, cruzada, cruzado, murrah, mediterrâneo, jafarabadi, carabao

Accounts can add custom breeds with nicknames (e.g., "NEL" → "nelore").

**Examples:**

Single birth:
```
1111 m angus
```

Birth with area:
```
88888 m Cruzado
espirito santo
```

Multiple births with date:
```
15/01
12941 F Angus
12942 M Nelore
Filhos de Eva
```

---

### Calf Messages (Newborn without Tag)

Records a newborn calf by referencing its mother (dam) instead of a tag. Used when the calf hasn't been tagged yet.

**Format:**
```
calf {dam} {sex} {breed}
{area}
```

**Fields:**
- `calf` - Keyword to indicate this is a calf entry. Also accepts: `bezerro`, `bezerra`, `bez`
- `dam` - Mother's ear tag number (required, must be > 0)
- `sex` - `m` or `f` (case insensitive)
- `breed` - Must match a known breed or account-specific breed nickname
- `area` - Optional, on a separate line
- `date` - Optional, format `dd/mm` on any line

Creates a birth record with `tag: 0` and the dam's tag stored in the `dam` field.

**Examples:**

Single calf:
```
calf 12345 f nelore
```

Using Portuguese keywords:
```
bezerro 12345 m nelore
bezerra 12345 f angus
bez 12345 f brangus
```

Multiple calves from same dam:
```
calf 12345 f nelore
calf 12345 m nelore
espirito santo
```

Mixed tagged births and calves:
```
15/01
88888 m angus
calf 12345 f nelore
calf 67890 m brangus
Fazenda Norte
```

---

### Death Messages

Updates an existing animal's status to deceased. Looks up the animal by tag number.

**Format:**
```
{tag} {cause}
```

**Fields:**
- `tag` - Numeric ear tag of existing animal
- `cause` - One of: `morreu`, `morto`, `nasceu morto`, `aborto`, `natimorto`, `natimortos`
- `date` - Optional, format `dd/mm` on any line

**Examples:**

Single death:
```
1234 morreu
```

Multiple deaths with date:
```
10/02
5678 aborto
9999 natimorto
```

---

### Rain Messages

Records rainfall measurements.

**Format:**
```
{day}/{month} {amount}mm
```

**Fields:**
- `day/month` - Date in dd/mm format
- `amount` - Rainfall in millimeters
- `mm` - Unit indicator (can have space before: `25mm` or `25 mm`)

**Examples:**

Single entry:
```
15/02 25mm
```

Multiple days:
```
14/02 12mm
15/02 25 mm
16/02 8mm
```

---

### Temperature Messages

Records temperature measurements.

**Format:**
```
{day}/{month} {temperature}c
```

**Fields:**
- `day/month` - Date in dd/mm format
- `temperature` - Temperature in Celsius
- `c` - Unit indicator (can have space before: `35c` or `35 c`, case insensitive)

**Examples:**

Single entry:
```
15/02 35c
```

Multiple days:
```
14/02 32 C
15/02 35c
16/02 28c
```

---

## Message Processing Notes

1. **Line Parsing**: Messages are split by newlines. Each line is checked against all parsers.

2. **Parser Priority**: Parsers are checked in order: Death, Birth, Rain, Temperature, Weather. A message matches only one parser.

3. **Date Handling**: If a date (`dd/mm`) is included in the message, it overrides the message timestamp. Dates use current year.

4. **Area Detection**: For birth messages, if no known area is found and the last line wasn't parsed as data, it's treated as a new area name.

5. **Breed Matching**: Breeds are matched against both system defaults and account-specific breeds (including nicknames).

6. **Multi-tenancy**: Phone numbers are mapped to accounts via the `teams` collection. All data is scoped to the sender's account.

## API Endpoints

See `CLAUDE.md` for full API documentation.

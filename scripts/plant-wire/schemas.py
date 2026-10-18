import copy, json, os, sys
out = sys.argv[1]
os.makedirs(out, exist_ok=True)
ident = {"type": "string", "pattern": "^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$"}
digest = {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"}
u64 = {"type": "integer", "minimum": 0}
i64 = {"type": "integer"}
desc = {"type": "string", "maxLength": 200}
quality = {"enum": ["good", "warming_up", "suspect", "out_of_range", "stale", "disconnected", "manual_override", "out_of_service"]}
value = {"type": ["number", "integer", "boolean", "string"], "maxLength": 256}
data_type = {"enum": ["number", "integer", "boolean", "enum", "string"]}
rng = {"type": "object", "required": ["min", "max"], "properties": {"min": {"type": "number"}, "max": {"type": "number"}}}
reasons = ["malformed", "unknown_device", "unknown_command", "out_of_range", "expired", "wrong_boot", "duplicate", "stale_epoch",
           "policy_not_allowed", "interlocked", "not_ready", "local_control", "permissive_not_met", "locked_out", "maintenance_isolated", "session_ending"]

def envelope(rtype, lane, required, props, desc_text):
    base = {"type": {"const": rtype}, "protocol_version": {"type": "integer", "minimum": 1}, "session_id": ident, "seq": u64,
            "lane": {"const": lane}}
    base.update(props)
    return {"$schema": "https://json-schema.org/draft/2020-12/schema", "$id": "urn:plant-wire:v1:" + rtype,
            "title": rtype, "description": desc_text, "type": "object",
            "required": ["type", "protocol_version", "session_id", "seq", "lane"] + required,
            "properties": base}

metric_template = {"type": "object", "required": ["name", "data_type"], "properties": {
    "name": ident, "data_type": data_type, "unit": {"type": "string", "maxLength": 32}, "eu_range": rng, "instrument_range": rng,
    "resolution": {"type": "number", "exclusiveMinimum": 0}, "enum_values": {"type": "array", "maxItems": 64, "items": {"type": "string", "maxLength": 64}},
    "scan_class": {"type": "string", "maxLength": 32}, "deadband": {"type": "number", "minimum": 0}, "description": desc, "writable": {"type": "boolean"},
    "path": {"type": "string", "maxLength": 256}}}
command_template = {"type": "object", "required": ["name", "args_schema"], "properties": {
    "name": ident, "args_schema": {"type": "object"}, "risk_class": {"enum": ["low", "medium", "high", "critical"]}, "safe_stop": {"type": "boolean"}, "description": desc}}
kind = {"type": "object", "required": ["kind", "metrics", "commands"], "properties": {
    "kind": ident, "metrics": {"type": "array", "maxItems": 512, "items": metric_template}, "commands": {"type": "array", "maxItems": 128, "items": command_template}}}
kind_entry = copy.deepcopy(kind)
kind_entry["required"] = kind["required"] + ["kind_digest"]
kind_entry["properties"]["kind_digest"] = digest
kind_file = {"$schema": "https://json-schema.org/draft/2020-12/schema", "$id": "urn:plant-wire:v1:KIND", "title": "KIND",
             "description": "A kind file: exactly one NBIRTH.kinds entry without its kind_digest, which a kind file must not carry (it is computed from the file). "
                            "NBIRTH.kinds items are this definition plus a required kind_digest. A kind file is the contract: a simulated implementation is "
                            "proven equal to it by the implements check, and a device is proven to match it only by announcing a kind_digest equal to the file's. "
                            "In a metric, a range needs min below max (checked by the simulator, which a schema cannot express).",
             **copy.deepcopy(kind)}
kind_file["properties"]["kind_digest"] = {"not": {}, "description": "Forbidden in a kind file: the digest is computed from the file."}
board = {"type": "object", "required": ["board", "environment", "members"], "properties": {"board": ident,
    "environment": {"enum": ["simulated", "hardware"], "description": "What this board is: the origin of every device born on it. A claim the host checks against the node's environment (SPEC section 2), never inferred."},
    "members": {"type": "array", "maxItems": 4096, "items": ident}}}
current = {"type": "object", "required": ["name", "alias", "value", "quality"], "properties": {"name": ident, "alias": u64, "value": value, "quality": quality}}
born = {"type": "object", "required": ["name", "alias", "value", "quality"], "properties": {"name": ident, "alias": u64,
    "value": {"type": ["number", "integer", "boolean", "string", "null"], "maxLength": 256}, "quality": quality},
    "if": {"properties": {"value": {"type": "null"}}, "required": ["value"]}, "then": {"properties": {"quality": {"const": "warming_up"}}}}
sample = {"type": "object", "required": ["alias", "value", "quality", "event_time_ns", "time_source"], "properties": {
    "alias": u64, "value": value, "quality": quality, "event_time_ns": i64, "time_source": {"enum": ["device", "gateway", "receipt"]},
    "uncertainty_ms": {"type": "number", "minimum": 0}, "device_mono_us": u64}}
evidence = {"type": "object", "additionalProperties": False,
    "required": ["command_id", "device", "boot_id", "metric", "outcome", "expected", "tolerance"],
    "properties": {"command_id": ident, "device": ident, "boot_id": ident, "metric": ident,
        "outcome": {"enum": ["observed", "not_observed"]}, "expected": {"type": "number"},
        "tolerance": {"type": "number", "minimum": 0}, "observed_value": {"type": "number"}},
    "if": {"properties": {"outcome": {"const": "observed"}}},
    "then": {"required": ["observed_value"]}}

schemas = {
 "KIND": kind_file,
 "NBIRTH": envelope("NBIRTH", "control", ["protocol_min", "protocol_max", "contract_version", "contract_digest", "environment", "bd_seq", "prior_last_seq", "capability_digest", "kinds", "boards"], {
    "protocol_min": {"type": "integer", "minimum": 1}, "protocol_max": {"type": "integer", "minimum": 1},
    "contract_version": {"type": "string", "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$"}, "contract_digest": digest,
    "environment": {"enum": ["simulated", "hardware", "mixed"], "description": "The node's environment. simulated: every board is simulated; hardware: every board is hardware; mixed: at least one board of each. A schema cannot say that, so a host checks it at the handshake (SPEC section 2)."}, "bd_seq": u64, "prior_last_seq": {"type": "integer", "minimum": -1},
    "capability_digest": digest, "kinds": {"type": "array", "maxItems": 256, "items": kind_entry}, "boards": {"type": "array", "maxItems": 4096, "items": board},
    "node_metrics": {"type": "object", "properties": {"clock_mode": {"enum": ["virtual", "wall"]}, "run_state": {"enum": ["running", "paused", "ended"]}}}},
    "Node birth: the session start or a rebirth. Declares each kind once and the environment of each board; grants nothing."),
 "DBIRTH": envelope("DBIRTH", "data", ["device", "kind", "board", "boot_id", "metadata_version", "birth_time_ns", "metrics"], {
    "device": ident, "kind": ident, "board": ident, "boot_id": ident, "metadata_version": u64, "birth_time_ns": i64,
    "overrides": {"type": "object", "properties": {"add": {"type": "array", "items": metric_template}, "narrow": {"type": "array", "items": {"type": "object", "required": ["name"], "properties": {"name": ident, "eu_range": rng}}}}},
    "metrics": {"type": "array", "maxItems": 512, "items": born}},
    "Device birth: kind, board, boot identity and the current value of every metric (null only while warming up)."),
 "DDATA": envelope("DDATA", "data", ["device", "boot_id", "historical", "metrics"], {
    "device": ident, "boot_id": ident, "historical": {"type": "boolean"}, "metrics": {"type": "array", "minItems": 1, "maxItems": 500, "items": sample}},
    "Device data: one or more samples. Historical data never updates live state or fires real-time rules."),
 "DDEATH": envelope("DDEATH", "control", ["device", "boot_id", "reason", "last_seq"], {
    "device": ident, "boot_id": ident, "reason": {"enum": ["offline", "link_down", "retired", "reboot", "poll_failure"]}, "last_seq": u64},
    "Device death: the whole device went offline. A failed sensor inside a live device is a quality code."),
 "NDEATH": envelope("NDEATH", "control", ["bd_seq", "reason"], {"bd_seq": u64, "reason": {"enum": ["shutdown", "connection_lost"]}},
    "Node death: graceful end. On socket loss the host synthesises it; a stale bd_seq is ignored."),
 "DRECEIPT": envelope("DRECEIPT", "control", ["command_id", "status", "boot_id"], {
    "command_id": ident, "status": {"enum": ["accepted", "refused", "queued"]}, "reason": {"enum": reasons}, "boot_id": ident},
    "Command receipt: accepted, refused (with a reason class) or queued."),
 "DRESULT": envelope("DRESULT", "control", ["command_id", "reported_outcome", "boot_id"], {
    "command_id": ident, "reported_outcome": {"enum": ["succeeded", "failed", "unknown"]}, "boot_id": ident},
    "Command result as the device reports it. Self-reported: it may disagree with reality."),
 "DSTATE": envelope("DSTATE", "control", ["device", "boot_id", "self_reported", "safe_state", "outputs", "values"], {
    "device": ident, "boot_id": ident, "self_reported": {"const": True}, "safe_state": {"type": "boolean"},
    "outputs": {"type": "array", "items": {"type": "object", "required": ["name", "value"], "properties": {"name": ident, "value": value}}},
    "values": {"type": "array", "items": current}},
    "Device state for reconciliation. Always self-reported; safety-relevant verification uses independent feedback."),
 "NSYNC": envelope("NSYNC", "control", ["through_seq", "world_time_ns"], {"through_seq": u64, "world_time_ns": i64},
    "Barrier: the node has sent everything up to through_seq for the slice ending at world_time_ns and waits for HSYNC."),
 "STATE": envelope("STATE", "control", ["host_id", "online", "heartbeat", "heartbeat_interval_ms", "authority_epoch", "protocol_min", "protocol_max"], {
    "host_id": ident, "online": {"type": "boolean"}, "heartbeat": u64, "heartbeat_interval_ms": {"type": "integer", "minimum": 100, "maximum": 600000},
    "authority_epoch": u64, "protocol_min": {"type": "integer", "minimum": 1}, "protocol_max": {"type": "integer", "minimum": 1},
    "refusal": {"enum": ["version_unsupported", "contract_unsupported", "environment_mismatch", "inventory_mismatch", "protocol_invalid"]}},
    "Host state and heartbeat; with online false and a refusal, the host refuses the session. online false is otherwise a graceful goodbye."),
 "NCMD": envelope("NCMD", "control", ["command"], {"command": {"enum": ["rebirth"]}},
    "Node command: rebirth republishes every birth before further data."),
 "DCMD": envelope("DCMD", "control", ["command_id", "idempotency_key", "device", "command", "args", "expected_boot_id", "deadline_ns", "not_before_ns", "authority_epoch", "policy_digest", "fallback_on_lease_loss"], {
    "command_id": ident, "idempotency_key": digest, "device": ident, "command": ident, "args": {"type": "object"},
    "expected_boot_id": ident, "deadline_ns": u64, "not_before_ns": u64, "authority_epoch": u64, "policy_digest": digest,
    "fallback_on_lease_loss": {"enum": ["revert_to_local", "hold_last", "safe_off"]}},
    "Device command with its safety fields: absolute deadline, boot identity, authority epoch and policy digest."),
 "DQUERY": envelope("DQUERY", "control", ["device"], {"device": ident}, "State query for reconciliation."),
 "HSYNC": envelope("HSYNC", "control", ["through_seq"], {"through_seq": u64,
    "command_evidence": {"type": "array", "maxItems": 128, "items": evidence}},
    "Barrier reply: the host has processed every node record up to through_seq and sent its commands for that slice."),
}
schemas["DCMD"]["additionalProperties"] = False
schemas["DRECEIPT"]["if"] = {"properties": {"status": {"const": "refused"}}, "required": ["status"]}
schemas["DRECEIPT"]["then"] = {"required": ["reason"]}
schemas["DRECEIPT"]["else"] = {"not": {"required": ["reason"]}}
for name, schema in schemas.items():
    with open(os.path.join(out, name + ".schema.json"), "w") as f:
        json.dump(schema, f, indent=2)
        f.write("\n")
print(len(schemas))

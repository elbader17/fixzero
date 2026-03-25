# Comparación: fixzero vs QuickFIX/Go

## 📋 Funcionalidades QuickFIX/Go que NO tiene fixzero

### 1. **Gestión de Sesiones (Session Management)**
- Manejo automático de secuencia de mensajes
- Reconnection automática
- Heartbeat y heartbeats faltantes
- Resend requests (resend de mensajes perdidos)
- Validación de sequence numbers

### 2. **Conectores de Red**
- `Initiator` - Cliente que inicia conexión
- `Acceptor` - Servidor que acepta conexiones
- Soporte TLS/SSL integrado
- Soporte para proxy protocol

### 3. **Validación de Mensajes**
- Data dictionary (especificaciones FIX)
- Validación de campos requeridos
- Validación de tipos de datos
- Validación de valores enum
- Repeating groups validation

### 4. **Almacenamiento de Estado**
- `MessageStore` interface
- Implementaciones: Memory, SQLite, file-based
- Persistencia de secuencia de mensajes
- Recovery después de restart

### 5. **Logging**
- `Log` interface
- Implementaciones: File, Screen, custom
- Logging de mensajes y eventos

### 6. **Message Router**
- Routing de mensajes por tipo
- handlers por message type

### 7. **Codegen de Mensajes**
- Código generado desde specs XML
- Tipos fuertemente tipados (NewOrderSingle, ExecutionReport, etc.)
- Enums para valores válidos

### 8. **Repeating Groups**
- Soporte completo para grupos repetitivos
- Validación automática de grupo

---

## ✅ Lo que SÍ tiene fixzero

| Funcionalidad | Status |
|--------------|--------|
| Parse de mensajes | ✅ |
| Serialize de mensajes | ✅ |
| Acceso a campos | ✅ |
| Zero-allocation | ✅ |
| Pool de memoria | ✅ |
| Buffer pools | ✅ |
| Parser reusables | ✅ |
| Builder API | ✅ |
| Tags estándar | ✅ |
| Tipos de mensaje estándar | ✅ |

---

## 🎯 Cuándo usar cada librería

### Usar **QuickFIX/Go** cuando:
- Necesitas implementación completa de protocolo
- Requiere gestión de sesiones automáticamente
- No tienes requisitos extremos de latencia
- Necesitas validación de mensajes
- Usas spec-driven development
- Necesitas persistencia de estado

### Usar **fixzero** cuando:
- Latencia ultra-baja es crítica (HFT)
- Controlas todo el flujo de mensajes
- Ya tienes tu propia gestión de sesiones
- Necesitas máximo rendimiento
- Tienes infraestructura de trading existente

---

## 📊 Trade-off

| Aspecto | QuickFIX/Go | fixzero |
|---------|-------------|---------|
| Latencia | Alta (3-9µs) | Ultra-baja (115ns) |
| Allocations | 20-56 por msg | 0-2 por msg |
| Funcionalidad | Completa | Core only |
| Complejidad | Alta | Baja |
| Overhead | Alto | Mínimo |
| Control | Bajo | Alto |


# Decisiones de diseño — Mesa de Ayuda

## D1 · Orden de los middleware

- **Decisión:** Registro afuera, Recuperación adentro (`r.Use(middleware.Registro); r.Use(middleware.Recuperacion)`).
- **Por qué:** queremos que toda petición quede registrada en el log, incluidas las que terminan en pánico. Con este orden, Registro sigue vivo por fuera cuando ocurre un pánico, porque Recuperación ya lo atrapó antes de que suba más allá.
- **Alternativa descartada:** Recuperación afuera, Registro adentro. Con ese orden, el pánico salta por encima de Registro (que no tiene su propio `recover()`), y esa petición nunca queda anotada en el log — el servidor responde 500 pero el registro no se entera.
- **Cómo lo comprobamos:** con `GET /explotar` en ambos órdenes. En el primer caso solo apareció la línea del PÁNICO; en el segundo aparecieron ambas líneas (PÁNICO y el registro con el 500).

## D2 · Código para JSON roto frente a datos inválidos

- **Decisión:** JSON roto → 400 (`cuerpo_malformado`). JSON válido pero que viola una regla de negocio (título corto, prioridad inválida) → 422 (`datos_invalidos`).
- **Por qué:** son dos tipos de error distintos para el cliente. Un 400 significa "no armaste bien la petición" (problema de forma/sintaxis); un 422 significa "la petición está bien armada, pero el dato en sí no cumple una regla" (problema semántico). Separarlos le da al cliente información más precisa sobre qué corregir.
- **Alternativa descartada:** usar 400 para ambos casos. Es más simple, pero mezcla dos causas de error distintas bajo el mismo código, y le quita al cliente la posibilidad de distinguir "mi código de programación está mal" de "el usuario ingresó un dato inválido".
- **Cómo lo comprobamos:** con tres peticiones POST /tickets: una con JSON roto (`400`), una con título corto (`422`), y una con prioridad inválida (`422`).
## D3 · (pendiente — la eligen ustedes)
## D3 · Campos desconocidos en el body

- **Decisión:** rechazar el body si trae un campo que `entradaCrear` no reconoce (`titulo`, `prioridad`), usando `dec.DisallowUnknownFields()`.
- **Por qué:** para que el cliente sepa exactamente qué pasó con su petición en vez de que su intento se ignore en silencio. Si alguien manda un campo como `estado` que no le corresponde fijar al crear, prefiero avisarle con un error claro en vez de que el servidor lo descarte sin decir nada.
- **Alternativa descartada:** ignorar los campos desconocidos (comportamiento por defecto de Go). Es más tolerante, pero deja al cliente sin saber si su dato se aplicó o no.
- **Cómo lo comprobamos:** POST /tickets con `{"titulo":"Pantalla rota","prioridad":"alta","estado":"cerrado"}` → debería dar 400, `cuerpo_malformado`.
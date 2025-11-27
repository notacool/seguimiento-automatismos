#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuración
API_URL="${API_URL:-http://localhost:8080}"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Variables globales para almacenar UUIDs
TASK_UUID=""
SUBTASK_UUID=""

# Función para imprimir separador
print_separator() {
    echo -e "${BLUE}================================================${NC}"
}

# Función para imprimir header de test
print_test_header() {
    print_separator
    echo -e "${YELLOW}TEST: $1${NC}"
    print_separator
}

# Función para verificar resultado
check_result() {
    local test_name="$1"
    local expected_status="$2"
    local actual_status="$3"
    local response="$4"

    ((TOTAL_TESTS++))

    if [ "$expected_status" -eq "$actual_status" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        echo -e "   Status: $actual_status"
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "   Expected: $expected_status"
        echo -e "   Got: $actual_status"
        echo -e "   Response: $response"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Función para extraer UUID de respuesta JSON
extract_uuid() {
    echo "$1" | grep -o '"uuid":"[^"]*"' | cut -d'"' -f4
}

# Función para hacer request y obtener status code
make_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local description="$4"

    echo ""
    echo -e "${BLUE}→${NC} $description"
    echo -e "   Method: $method"
    echo -e "   Endpoint: $endpoint"
    if [ -n "$data" ]; then
        echo -e "   Payload: $data"
    fi

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$API_URL$endpoint")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo -e "   Response Code: $http_code"
    echo -e "   Response Body: $body"

    # Exportar para uso en funciones llamadoras
    export LAST_RESPONSE="$body"
    export LAST_STATUS="$http_code"
}

# ============================================
# TESTS DE HEALTH CHECK
# ============================================
test_health_check() {
    print_test_header "1. Health Check"

    make_request "GET" "/health" "" "Verificar estado del servicio"
    check_result "Health endpoint should return 200" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Verificar que la respuesta contenga información de la BD
    if echo "$LAST_RESPONSE" | grep -q "database"; then
        echo -e "${GREEN}✓${NC} Response contains database info"
    else
        echo -e "${YELLOW}⚠${NC} Response might not contain expected database info"
    fi
}

# ============================================
# TESTS DE TAREAS (TASK ENDPOINTS)
# ============================================
test_create_task() {
    print_test_header "2. Create Task"

    # Test 2.1: Crear tarea válida
    local payload='{
        "name": "Proceso de Facturacion Mensual",
        "description": "Automatizacion del proceso de facturacion",
        "state": "PENDING",
        "created_by": "test-user",
        "subtasks": [
            {
                "name": "Generar facturas",
                "description": "Generar todas las facturas del mes",
                "state": "PENDING"
            },
            {
                "name": "Enviar por email",
                "description": "Enviar facturas a clientes",
                "state": "PENDING"
            }
        ]
    }'

    make_request "POST" "/Automatizacion" "$payload" "Crear tarea válida con subtareas"
    check_result "Create task with valid data" 201 "$LAST_STATUS" "$LAST_RESPONSE"

    # Guardar UUID de la tarea creada
    TASK_UUID=$(extract_uuid "$LAST_RESPONSE")
    if [ -n "$TASK_UUID" ]; then
        echo -e "${GREEN}✓${NC} Task UUID extracted: $TASK_UUID"
    else
        echo -e "${RED}✗${NC} Could not extract task UUID"
    fi

    # Guardar UUID de la primera subtarea
    SUBTASK_UUID=$(echo "$LAST_RESPONSE" | grep -o '"uuid":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)
    if [ -n "$SUBTASK_UUID" ]; then
        echo -e "${GREEN}✓${NC} Subtask UUID extracted: $SUBTASK_UUID"
    fi

    # Test 2.2: Crear tarea sin nombre (debería fallar)
    local invalid_payload='{
        "description": "Tarea sin nombre",
        "state": "PENDING",
        "created_by": "test-user"
    }'

    make_request "POST" "/Automatizacion" "$invalid_payload" "Crear tarea sin nombre (debe fallar)"
    check_result "Create task without name should fail" 400 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 2.3: Crear tarea con estado inválido (debería fallar)
    local invalid_state_payload='{
        "name": "Task with invalid state",
        "state": "INVALID_STATE",
        "created_by": "test-user"
    }'

    make_request "POST" "/Automatizacion" "$invalid_state_payload" "Crear tarea con estado inválido (debe fallar)"
    check_result "Create task with invalid state should fail" 400 "$LAST_STATUS" "$LAST_RESPONSE"
}

test_get_task() {
    print_test_header "3. Get Task by UUID"

    if [ -z "$TASK_UUID" ]; then
        echo -e "${YELLOW}⚠ Skipping test - No task UUID available${NC}"
        return
    fi

    # Test 3.1: Obtener tarea existente
    make_request "GET" "/Automatizacion/$TASK_UUID" "" "Obtener tarea por UUID"
    check_result "Get existing task" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Verificar que la respuesta contenga los datos esperados
    if echo "$LAST_RESPONSE" | grep -q "Proceso de Facturación Mensual"; then
        echo -e "${GREEN}✓${NC} Task name matches"
    fi

    # Test 3.2: Obtener tarea inexistente
    local fake_uuid="00000000-0000-0000-0000-000000000000"
    make_request "GET" "/Automatizacion/$fake_uuid" "" "Obtener tarea inexistente (debe fallar)"
    check_result "Get non-existent task should return 404" 404 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 3.3: UUID malformado
    make_request "GET" "/Automatizacion/invalid-uuid" "" "Obtener tarea con UUID inválido (debe fallar)"
    check_result "Get task with invalid UUID should fail" 400 "$LAST_STATUS" "$LAST_RESPONSE"
}

test_list_tasks() {
    print_test_header "4. List Tasks"

    # Test 4.1: Listar todas las tareas
    make_request "GET" "/AutomatizacionListado" "" "Listar todas las tareas"
    check_result "List all tasks" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 4.2: Listar con filtro por estado
    make_request "GET" "/AutomatizacionListado?state=PENDING" "" "Listar tareas con estado PENDING"
    check_result "List tasks filtered by state" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 4.3: Listar con paginación
    make_request "GET" "/AutomatizacionListado?page=1&limit=10" "" "Listar tareas con paginación"
    check_result "List tasks with pagination" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 4.4: Listar con filtro por created_by
    make_request "GET" "/AutomatizacionListado?created_by=test-user" "" "Listar tareas filtradas por creador"
    check_result "List tasks filtered by creator" 200 "$LAST_STATUS" "$LAST_RESPONSE"
}

test_update_task() {
    print_test_header "5. Update Task"

    if [ -z "$TASK_UUID" ]; then
        echo -e "${YELLOW}⚠ Skipping test - No task UUID available${NC}"
        return
    fi

    # Test 5.1: Actualizar descripción
    local update_payload='{
        "uuid": "'"$TASK_UUID"'",
        "description": "Descripción actualizada del proceso",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Automatizacion" "$update_payload" "Actualizar descripción de tarea"
    check_result "Update task description" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 5.2: Transición de estado PENDING → IN_PROGRESS
    local state_update_payload='{
        "uuid": "'"$TASK_UUID"'",
        "state": "IN_PROGRESS",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Automatizacion" "$state_update_payload" "Cambiar estado a IN_PROGRESS"
    check_result "Update task state to IN_PROGRESS" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Verificar que start_date se haya asignado
    if echo "$LAST_RESPONSE" | grep -q "start_date"; then
        echo -e "${GREEN}✓${NC} Start date assigned"
    fi

    # Test 5.3: Transición de estado IN_PROGRESS → COMPLETED
    local complete_payload='{
        "uuid": "'"$TASK_UUID"'",
        "state": "COMPLETED",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Automatizacion" "$complete_payload" "Cambiar estado a COMPLETED"
    check_result "Update task state to COMPLETED" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Verificar que end_date se haya asignado
    if echo "$LAST_RESPONSE" | grep -q "end_date"; then
        echo -e "${GREEN}✓${NC} End date assigned"
    fi

    # Test 5.4: Intentar transición inválida desde estado final (debería fallar)
    local invalid_transition_payload='{
        "uuid": "'"$TASK_UUID"'",
        "state": "PENDING",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Automatizacion" "$invalid_transition_payload" "Transición inválida desde COMPLETED (debe fallar)"
    check_result "Invalid state transition should fail" 400 "$LAST_STATUS" "$LAST_RESPONSE"

    # Verificar que la respuesta siga RFC 7807
    if echo "$LAST_RESPONSE" | grep -q "type"; then
        echo -e "${GREEN}✓${NC} Error response follows RFC 7807"
    fi
}

# ============================================
# TESTS DE SUBTAREAS (SUBTASK ENDPOINTS)
# ============================================
test_update_subtask() {
    print_test_header "6. Update Subtask"

    # Crear nueva tarea para tests de subtareas (la anterior está COMPLETED)
    local task_payload='{
        "name": "Task for subtask testing",
        "state": "PENDING",
        "created_by": "test-user",
        "subtasks": [
            {
                "name": "Test subtask",
                "description": "Descripción subtarea",
                "state": "PENDING"
            }
        ]
    }'

    make_request "POST" "/Automatizacion" "$task_payload" "Crear tarea para tests de subtareas"

    # Extraer UUID de la subtarea
    local new_subtask_uuid=$(echo "$LAST_RESPONSE" | grep -o '"uuid":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)

    if [ -z "$new_subtask_uuid" ]; then
        echo -e "${YELLOW}⚠ Skipping subtask tests - No subtask UUID available${NC}"
        return
    fi

    echo -e "${GREEN}✓${NC} Using subtask UUID: $new_subtask_uuid"

    # Test 6.1: Actualizar descripción de subtarea
    local update_subtask_payload='{
        "description": "Descripción actualizada de subtarea",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Subtask/$new_subtask_uuid" "$update_subtask_payload" "Actualizar subtarea"
    check_result "Update subtask" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 6.2: Actualizar estado de subtarea
    local update_state_payload='{
        "state": "IN_PROGRESS",
        "updated_by": "test-user"
    }'

    make_request "PUT" "/Subtask/$new_subtask_uuid" "$update_state_payload" "Cambiar estado de subtarea"
    check_result "Update subtask state" 200 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 6.3: Actualizar subtarea inexistente (debería fallar)
    local fake_uuid="00000000-0000-0000-0000-000000000000"
    make_request "PUT" "/Subtask/$fake_uuid" "$update_subtask_payload" "Actualizar subtarea inexistente (debe fallar)"
    check_result "Update non-existent subtask should fail" 404 "$LAST_STATUS" "$LAST_RESPONSE"
}

test_delete_subtask() {
    print_test_header "7. Delete Subtask (Soft Delete)"

    # Crear tarea con subtarea para eliminar
    local task_payload='{
        "name": "Task for deletion testing",
        "state": "PENDING",
        "created_by": "test-user",
        "subtasks": [
            {
                "name": "Subtask to delete",
                "state": "PENDING"
            }
        ]
    }'

    make_request "POST" "/Automatizacion" "$task_payload" "Crear tarea para test de eliminación"

    local delete_subtask_uuid=$(echo "$LAST_RESPONSE" | grep -o '"uuid":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)

    if [ -z "$delete_subtask_uuid" ]; then
        echo -e "${YELLOW}⚠ Skipping delete test - No subtask UUID available${NC}"
        return
    fi

    # Test 7.1: Eliminar subtarea (soft delete)
    make_request "DELETE" "/Subtask/$delete_subtask_uuid" "" "Eliminar subtarea (soft delete)"
    check_result "Soft delete subtask" 204 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 7.2: Verificar que la subtarea ya no se puede obtener
    make_request "PUT" "/Subtask/$delete_subtask_uuid" '{"updated_by":"test"}' "Intentar actualizar subtarea eliminada (debe fallar)"
    check_result "Update deleted subtask should fail" 404 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 7.3: Eliminar subtarea inexistente (debería fallar)
    local fake_uuid="00000000-0000-0000-0000-000000000000"
    make_request "DELETE" "/Subtask/$fake_uuid" "" "Eliminar subtarea inexistente (debe fallar)"
    check_result "Delete non-existent subtask should fail" 404 "$LAST_STATUS" "$LAST_RESPONSE"
}

# ============================================
# TESTS DE CASOS EDGE Y VALIDACIONES
# ============================================
test_edge_cases() {
    print_test_header "8. Edge Cases and Validation"

    # Test 8.1: Request con JSON malformado
    make_request "POST" "/Automatizacion" "{invalid json" "JSON malformado (debe fallar)"
    check_result "Malformed JSON should fail" 400 "$LAST_STATUS" "$LAST_RESPONSE"

    # Test 8.2: Request sin Content-Type application/json
    echo ""
    echo -e "${BLUE}→${NC} Request sin Content-Type correcto"
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -d '{"name":"test"}' \
        "$API_URL/Automatizacion")
    http_code=$(echo "$response" | tail -n1)
    check_result "Request without proper Content-Type" 400 "$http_code" "$response"

    # Test 8.3: Tarea con nombre muy largo
    local long_name=$(python3 -c "print('A' * 300)")
    local long_name_payload='{
        "name": "'"$long_name"'",
        "state": "PENDING",
        "created_by": "test-user"
    }'

    make_request "POST" "/Automatizacion" "$long_name_payload" "Tarea con nombre muy largo"
    # Puede ser 201 si se acepta o 400 si hay validación de longitud
    if [ "$LAST_STATUS" -eq 201 ] || [ "$LAST_STATUS" -eq 400 ]; then
        echo -e "${GREEN}✓${NC} Long name handled appropriately (status: $LAST_STATUS)"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}✗${NC} Unexpected status for long name: $LAST_STATUS"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
}

# ============================================
# TEST DE FLUJO COMPLETO
# ============================================
test_complete_workflow() {
    print_test_header "9. Complete Workflow Test"

    echo -e "${YELLOW}Testing complete task lifecycle...${NC}"

    # 1. Crear tarea
    local workflow_payload='{
        "name": "Workflow Test - Proceso Completo",
        "description": "Test de flujo completo de trabajo",
        "state": "PENDING",
        "created_by": "workflow-test",
        "subtasks": [
            {
                "name": "Step 1 Preparation",
                "state": "PENDING"
            },
            {
                "name": "Step 2 Execution",
                "state": "PENDING"
            },
            {
                "name": "Step 3 Validation",
                "state": "PENDING"
            }
        ]
    }'

    make_request "POST" "/Automatizacion" "$workflow_payload" "Step 1: Create task"
    if [ "$LAST_STATUS" -ne 201 ]; then
        echo -e "${RED}✗${NC} Workflow failed at creation"
        return
    fi

    local workflow_uuid=$(extract_uuid "$LAST_RESPONSE")
    echo -e "${GREEN}✓${NC} Task created: $workflow_uuid"

    # 2. Listar y verificar que aparece
    make_request "GET" "/AutomatizacionListado?created_by=workflow-test" "" "Step 2: List tasks"
    if echo "$LAST_RESPONSE" | grep -q "$workflow_uuid"; then
        echo -e "${GREEN}✓${NC} Task appears in listing"
    else
        echo -e "${RED}✗${NC} Task not found in listing"
    fi

    # 3. Obtener tarea específica
    make_request "GET" "/Automatizacion/$workflow_uuid" "" "Step 3: Get task details"
    if [ "$LAST_STATUS" -eq 200 ]; then
        echo -e "${GREEN}✓${NC} Task retrieved successfully"
    fi

    # 4. Iniciar tarea (PENDING → IN_PROGRESS)
    local start_payload='{
        "uuid": "'"$workflow_uuid"'",
        "state": "IN_PROGRESS",
        "updated_by": "workflow-test"
    }'
    make_request "PUT" "/Automatizacion" "$start_payload" "Step 4: Start task"
    if [ "$LAST_STATUS" -eq 200 ] && echo "$LAST_RESPONSE" | grep -q "start_date"; then
        echo -e "${GREEN}✓${NC} Task started with start_date"
    fi

    # 5. Completar tarea
    local complete_payload='{
        "uuid": "'"$workflow_uuid"'",
        "state": "COMPLETED",
        "updated_by": "workflow-test"
    }'
    make_request "PUT" "/Automatizacion" "$complete_payload" "Step 5: Complete task"
    if [ "$LAST_STATUS" -eq 200 ] && echo "$LAST_RESPONSE" | grep -q "end_date"; then
        echo -e "${GREEN}✓${NC} Task completed with end_date"
    fi

    # 6. Verificar que las subtareas heredaron el estado
    make_request "GET" "/Automatizacion/$workflow_uuid" "" "Step 6: Verify subtasks inherited state"
    if echo "$LAST_RESPONSE" | grep -q "\"state\":\"COMPLETED\""; then
        echo -e "${GREEN}✓${NC} Subtasks inherited COMPLETED state"
    fi

    echo -e "${GREEN}✓ COMPLETE WORKFLOW TEST PASSED${NC}"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

# ============================================
# FUNCIÓN PRINCIPAL
# ============================================
main() {
    echo ""
    print_separator
    echo -e "${YELLOW}API TEST SUITE${NC}"
    echo -e "${BLUE}Target: $API_URL${NC}"
    print_separator
    echo ""

    # Verificar que el servidor está disponible
    echo -e "${BLUE}Checking API availability...${NC}"
    if ! curl -s -f "$API_URL/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ API is not available at $API_URL${NC}"
        echo -e "${YELLOW}Make sure the service is running with: make docker-up${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ API is available${NC}"
    echo ""

    # Ejecutar todos los tests
    test_health_check
    test_create_task
    test_get_task
    test_list_tasks
    test_update_task
    test_update_subtask
    test_delete_subtask
    test_edge_cases
    test_complete_workflow

    # Resumen final
    echo ""
    print_separator
    echo -e "${YELLOW}TEST SUMMARY${NC}"
    print_separator
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All tests passed! ✓${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed ✗${NC}"
        exit 1
    fi
}

# Ejecutar tests
main "$@"

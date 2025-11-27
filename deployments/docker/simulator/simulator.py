#!/usr/bin/env python3
"""
Simulador de Automatismo
Simula el ciclo de vida completo de automatizaciones para probar el sistema.
"""
import os
import time
import random
import requests
import logging
from typing import Dict, List, Optional
from datetime import datetime

# Configuración
API_BASE_URL = os.getenv('API_BASE_URL', 'http://api:8080')
SIMULATOR_INTERVAL = int(os.getenv('SIMULATOR_INTERVAL', '30'))  # segundos entre ciclos
SIMULATOR_NAME_PREFIX = os.getenv('SIMULATOR_NAME_PREFIX', 'Sim-Auto')
SIMULATOR_CREATED_BY = os.getenv('SIMULATOR_CREATED_BY', 'simulator-bot')

# Configurar logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class AutomatismoSimulator:
    """Simula automatismos que pasan por diferentes estados."""

    def __init__(self, api_url: str):
        self.api_url = api_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })
        self.active_tasks: Dict[str, Dict] = {}

    def health_check(self) -> bool:
        """Verifica que la API esté disponible."""
        try:
            response = self.session.get(f'{self.api_url}/health', timeout=5)
            if response.status_code == 200:
                data = response.json()
                logger.info(f"✓ API Health: {data.get('status')}, DB: {data.get('database')}")
                return True
            return False
        except Exception as e:
            logger.error(f"✗ Health check failed: {e}")
            return False

    def create_task(self, name: str, subtasks: Optional[List[Dict]] = None) -> Optional[str]:
        """Crea una nueva tarea."""
        try:
            payload = {
                'name': name,
                'created_by': SIMULATOR_CREATED_BY
            }
            if subtasks:
                payload['subtasks'] = subtasks

            response = self.session.post(
                f'{self.api_url}/Automatizacion',
                json=payload,
                timeout=10
            )

            if response.status_code == 201:
                task = response.json()
                task_id = task['id']
                logger.info(f"✓ Tarea creada: {name} (ID: {task_id})")
                return task_id
            else:
                logger.error(f"✗ Error al crear tarea: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"✗ Excepción al crear tarea: {e}")
            return None

    def update_task_state(self, task_id: str, state: str) -> bool:
        """Actualiza el estado de una tarea."""
        try:
            payload = {
                'id': task_id,
                'state': state,
                'updated_by': SIMULATOR_CREATED_BY
            }

            response = self.session.put(
                f'{self.api_url}/Automatizacion',
                json=payload,
                timeout=10
            )

            if response.status_code == 200:
                logger.info(f"✓ Tarea {task_id[:8]}... actualizada a {state}")
                return True
            else:
                logger.error(f"✗ Error al actualizar tarea: {response.status_code} - {response.text}")
                return False
        except Exception as e:
            logger.error(f"✗ Excepción al actualizar tarea: {e}")
            return False

    def update_subtask_state(self, subtask_id: str, state: str) -> bool:
        """Actualiza el estado de una subtarea."""
        try:
            payload = {
                'state': state,
                'updated_by': SIMULATOR_CREATED_BY
            }

            response = self.session.put(
                f'{self.api_url}/Subtask/{subtask_id}',
                json=payload,
                timeout=10
            )

            if response.status_code == 200:
                logger.info(f"✓ Subtarea {subtask_id[:8]}... actualizada a {state}")
                return True
            else:
                logger.error(f"✗ Error al actualizar subtarea: {response.status_code} - {response.text}")
                return False
        except Exception as e:
            logger.error(f"✗ Excepción al actualizar subtarea: {e}")
            return False

    def get_task(self, task_id: str) -> Optional[Dict]:
        """Obtiene una tarea por ID."""
        try:
            response = self.session.get(
                f'{self.api_url}/Automatizacion/{task_id}',
                timeout=10
            )

            if response.status_code == 200:
                return response.json()
            return None
        except Exception as e:
            logger.error(f"✗ Error al obtener tarea: {e}")
            return None

    def simulate_automatismo_cycle(self, task_id: str, task_name: str):
        """Simula el ciclo completo de un automatismo."""
        task = self.get_task(task_id)
        if not task:
            logger.warning(f"⚠ No se pudo obtener la tarea {task_id}")
            return

        current_state = task['state']
        subtasks = task.get('subtasks', [])

        # Lógica de transición de estados
        if current_state == 'PENDING':
            # Avanzar a IN_PROGRESS
            if self.update_task_state(task_id, 'IN_PROGRESS'):
                logger.info(f"🔄 {task_name}: PENDING → IN_PROGRESS")
                # Actualizar primera subtarea si existe
                if subtasks and len(subtasks) > 0:
                    first_subtask = subtasks[0]
                    if first_subtask['state'] == 'PENDING':
                        self.update_subtask_state(first_subtask['id'], 'IN_PROGRESS')

        elif current_state == 'IN_PROGRESS':
            # Decidir si completar o fallar (90% éxito, 10% fallo)
            if random.random() < 0.9:
                # Completar todas las subtareas primero
                task = self.get_task(task_id)
                if task:
                    for subtask in task.get('subtasks', []):
                        if subtask['state'] == 'PENDING':
                            self.update_subtask_state(subtask['id'], 'IN_PROGRESS')
                        if subtask['state'] == 'IN_PROGRESS':
                            self.update_subtask_state(subtask['id'], 'COMPLETED')
                
                # Completar la tarea
                if self.update_task_state(task_id, 'COMPLETED'):
                    logger.info(f"✅ {task_name}: IN_PROGRESS → COMPLETED")
                    # Marcar como completada para eliminarla del ciclo
                    if task_id in self.active_tasks:
                        del self.active_tasks[task_id]
            else:
                # Fallar la tarea
                if self.update_task_state(task_id, 'FAILED'):
                    logger.warning(f"❌ {task_name}: IN_PROGRESS → FAILED")
                    if task_id in self.active_tasks:
                        del self.active_tasks[task_id]

    def create_new_automatismo(self) -> Optional[str]:
        """Crea un nuevo automatismo con subtareas."""
        timestamp = datetime.now().strftime('%H%M%S')
        task_name = f"{SIMULATOR_NAME_PREFIX}-{timestamp}"

        # Crear subtareas
        subtasks = [
            {'name': 'Inicializacion', 'state': 'PENDING'},
            {'name': 'Procesamiento', 'state': 'PENDING'},
            {'name': 'Validacion', 'state': 'PENDING'},
            {'name': 'Finalizacion', 'state': 'PENDING'}
        ]

        task_id = self.create_task(task_name, subtasks=subtasks)
        if task_id:
            self.active_tasks[task_id] = {
                'name': task_name,
                'created_at': datetime.now()
            }
            logger.info(f"🆕 Nuevo automatismo creado: {task_name}")

        return task_id

    def run(self):
        """Ejecuta el simulador en bucle."""
        logger.info("=" * 60)
        logger.info("🤖 Simulador de Automatismos iniciado")
        logger.info(f"📍 API URL: {API_BASE_URL}")
        logger.info(f"⏱️  Intervalo: {SIMULATOR_INTERVAL} segundos")
        logger.info("=" * 60)

        # Esperar a que la API esté lista
        logger.info("⏳ Esperando a que la API esté disponible...")
        max_retries = 30
        retry = 0
        while retry < max_retries:
            if self.health_check():
                break
            retry += 1
            time.sleep(2)
        else:
            logger.error("✗ No se pudo conectar a la API después de varios intentos")
            return

        logger.info("✓ API disponible, iniciando simulación...")
        time.sleep(2)

        cycle_count = 0
        while True:
            try:
                cycle_count += 1
                logger.info(f"\n{'=' * 60}")
                logger.info(f"🔄 Ciclo #{cycle_count} - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
                logger.info(f"{'=' * 60}")

                # Crear nuevo automatismo ocasionalmente (30% de probabilidad)
                if len(self.active_tasks) < 5 and random.random() < 0.3:
                    self.create_new_automatismo()

                # Procesar automatismos activos
                if self.active_tasks:
                    logger.info(f"📊 Automatismos activos: {len(self.active_tasks)}")
                    for task_id, task_info in list(self.active_tasks.items()):
                        self.simulate_automatismo_cycle(task_id, task_info['name'])
                else:
                    logger.info("📊 No hay automatismos activos, creando uno nuevo...")
                    self.create_new_automatismo()

                # Esperar antes del siguiente ciclo
                logger.info(f"⏳ Esperando {SIMULATOR_INTERVAL} segundos hasta el siguiente ciclo...")
                time.sleep(SIMULATOR_INTERVAL)

            except KeyboardInterrupt:
                logger.info("\n🛑 Simulador detenido por el usuario")
                break
            except Exception as e:
                logger.error(f"✗ Error en el ciclo: {e}", exc_info=True)
                time.sleep(5)  # Esperar un poco antes de reintentar


if __name__ == '__main__':
    simulator = AutomatismoSimulator(API_BASE_URL)
    simulator.run()


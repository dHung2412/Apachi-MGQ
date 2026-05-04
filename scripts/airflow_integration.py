import json
import requests
import datetime
from airflow.models import Variable
from airflow.hooks.base import BaseHook

class DataCatalogHook:
    """
    Hook to push metadata to the Go Data Catalog API.
    """
    def __init__(self, conn_id='data_catalog_api'):
        # Get connection details from Airflow (Host, Port, Password as Token)
        self.connection = BaseHook.get_connection(conn_id)
        self.base_url = f"http://{self.connection.host}:{self.connection.port}"
        self.token = self.connection.password

    def ingest(self, payload):
        url = f"{self.base_url}/api/v1/ingest"
        headers = {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
        
        response = requests.post(url, headers=headers, json=payload)
        response.raise_for_status()
        return response.json()

def data_catalog_callback(context):
    """
    Example callback function to be used in on_success_callback.
    """
    ti = context['task_instance']
    dag_id = ti.dag_id
    task_id = ti.task_id
    run_id = context['run_id']
    
    # In a real scenario, you might extract these from Task return values (XCom)
    # or from the Operator's configuration.
    payload = {
        "event_type": "task_completion",
        "timestamp": datetime.datetime.utcnow().isoformat() + "Z",
        "source": "airflow",
        "dag_id": dag_id,
        "task_id": task_id,
        "run_id": run_id,
        "payload": {
            "dataset": {
                "name": "example_table",
                "urn": f"postgres:db.schema.{task_id}_output",
                "platform": "postgres"
            },
            "lineage": {
                "source_dataset": "raw_source_table",
                "target_dataset": f"{task_id}_output",
                "transform_type": "ETL_JOB",
                "mappings": [
                    {"source_column": "id", "target_column": "id", "transform": "copy"}
                ]
            }
        }
    }
    
    try:
        hook = DataCatalogHook()
        hook.ingest(payload)
        print(f"Successfully pushed metadata for {task_id}")
    except Exception as e:
        print(f"Failed to push metadata: {str(e)}")

# --- Usage in a DAG ---
# from airflow import DAG
# from airflow.operators.python import PythonOperator
#
# with DAG(..., on_success_callback=None) as dag:
#     task = PythonOperator(
#         task_id='my_data_task',
#         python_callable=my_func,
#         on_success_callback=data_catalog_callback  # <--- Bind here
#     )

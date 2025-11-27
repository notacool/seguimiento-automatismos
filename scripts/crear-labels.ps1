# Script para crear labels en GitHub usando gh
$labels = @(
    @{name = 'frontend'; description = 'Feature o bugfix del frontend'; color = '1abc9c' },
    @{name = 'backend'; description = 'Feature o bugfix del backend'; color = '3498db' },
    @{name = 'scope change'; description = 'Cambio de alcance del proyecto'; color = 'e67e22' },
    @{name = 'documentation'; description = 'Mejora o corrección de documentación'; color = 'f1c40f' },
    @{name = 'critical'; description = 'Issue crítico y prioritario'; color = 'e74c3c' },
    @{name = 'no prioritario'; description = 'Mejora no prioritaria'; color = '95a5a6' },
    @{name = 'scripts'; description = 'Cambios en scripts de apoyo'; color = '9b59b6' },
    @{name = 'linter'; description = 'Errores de linter'; color = '2ecc71' },
    @{name = 'analysis required'; description = 'Requiere análisis previo'; color = '34495e' },
    @{name = 'dependant'; description = 'Depende de otra issue'; color = '7f8c8d' },
    @{name = 'blocked'; description = 'Issue bloqueada'; color = 'c0392b' }
)
foreach ($label in $labels) {
    gh label create $label.name --description "$($label.description)" --color $label.color
}

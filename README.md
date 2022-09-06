# CRM deployment

* Main repos
    * crm-collection - for collection crm
        * app
        * db
        * rabbitmq
    * crm3-crm-collection-new - for base crm

* Deployments steps
    * Main CRM
        * MYSQl DB configure and run
        * run create-db.sql and damucrm.sql
        * run docker-compuse from main repo
        * port 8282
    * Collection
        * run create-db.sql and collection.sql
        * run docker-compose from the **app** directory
        * port 8283
